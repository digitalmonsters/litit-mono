package internal

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/kafka_listener/structs"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
	"io"
	"sync"
	"time"
)

var readerMutex sync.Mutex

type KafkaListenerLowLevelConfig struct {
	GroupId             string // can be empty
	ReadOnlyNewMessages bool
	Topic               string
	Hosts               string
	Tls                 bool
	KafkaAuth           boilerplate.KafkaAuth
	MinBytes            int
	MaxBytes            int
}

type KafkaListener struct {
	cfg               KafkaListenerLowLevelConfig
	ctx               context.Context
	readers           map[int]*kafka.Reader // key is partition; 0 - for GroupId
	targetTopic       string
	command           structs.ICommand
	listenerName      string
	cancelFn          context.CancelFunc
	hasRunningRequest bool
	dialer            *kafka.Dialer
}

func NewKafkaListener(config KafkaListenerLowLevelConfig, ctx context.Context, command structs.ICommand) *KafkaListener {
	if len(config.Topic) == 0 {
		panic("kafka topic should not be empty")
	}

	if config.MaxBytes == 0 {
		config.MaxBytes = 10e6 // 10 MB
	}

	dialer, err := GetKafkaDialer(config.Tls, config.KafkaAuth)

	if err != nil {
		panic(err)
	}

	localCtx, cancelFn := context.WithCancel(ctx)

	return &KafkaListener{
		cfg:          config,
		ctx:          localCtx,
		cancelFn:     cancelFn,
		targetTopic:  config.Topic,
		command:      command,
		dialer:       dialer,
		listenerName: fmt.Sprintf("kafka_listener_%v", config.Topic),
	}
}

func (k KafkaListener) GetTopic() string {
	return k.targetTopic
}

func (k *KafkaListener) getPartitionsForTopic() ([]int, error) {
	if len(k.cfg.GroupId) != 0 {
		return []int{0}, nil // 0 means that we dont care as we have GroupId
	}

	var finalPartitions []int

	for _, host := range boilerplate.SplitHostsToSlice(k.cfg.Hosts) {
		con, err := k.dialer.Dial("tcp", host)

		if err != nil {
			log.Err(err).Msgf("can not get connection to calculate partitions for topic %v", k.cfg.Topic)
			continue
		}

		partitions, err := con.ReadPartitions(k.cfg.Topic)

		if err != nil {
			log.Err(err).Msgf("can not get partitions for topic %v", k.cfg.Topic)
			_ = con.Close()
			continue
		}

		for _, p := range partitions {
			finalPartitions = append(finalPartitions, p.ID)
		}

		_ = con.Close()
	}

	if len(finalPartitions) == 0 {
		return nil, errors.New(fmt.Sprintf("no partitions found for topic %v", k.cfg.Topic))
	}

	return finalPartitions, nil
}

func (k *KafkaListener) getReaderForPartition(partition int) (*kafka.Reader, error) {
	readerMutex.Lock()
	defer readerMutex.Unlock()

	if v, ok := k.readers[partition]; ok {
		return v, nil
	}

	dialer, err := GetKafkaDialer(true, k.cfg.KafkaAuth)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	if len(k.cfg.GroupId) == 0 { // then we need to implement specific logic, as we dont have consumer group

	}

	var kafkaCfg = kafka.ReaderConfig{
		Brokers:        boilerplate.SplitHostsToSlice(k.cfg.Hosts),
		GroupID:        k.cfg.GroupId,
		Partition:      partition, // if GroupId
		Topic:          k.targetTopic,
		MinBytes:       k.cfg.MinBytes,
		MaxBytes:       k.cfg.MaxBytes,
		CommitInterval: time.Millisecond,
		Dialer:         dialer,
	}

	r := kafka.NewReader(kafkaCfg)

	k.readers[partition] = r

	return r, nil
}

func (k *KafkaListener) ListenInBatches(maxBatchSize int, maxDuration time.Duration) {
	var partitions []int
	var err error

	for k.ctx.Err() == nil {
		partitions, err = k.getPartitionsForTopic()

		if err != nil {
			log.Err(err).Send()

			time.Sleep(10 * time.Second)
		}

		break
	}

	for _, partition := range partitions {
		p := partition

		go func() {
			for k.ctx.Err() == nil {
				reader, err := k.getReaderForPartition(p)

				if err != nil {
					log.Err(err).Send()

					time.Sleep(10 * time.Second)
					continue
				}

				if err := k.listen(maxBatchSize, maxDuration, reader); err != nil {
					tx := apm_helper.StartNewApmTransaction(k.listenerName, "kafka_listener", nil, nil)

					apm_helper.CaptureApmError(err, tx)
					log.Err(err).Send()

					tx.End()
					time.Sleep(5 * time.Second)
				}
			}
		}()
	}
}

func (k *KafkaListener) closeReader(partitionId int) {
	readerMutex.Lock()
	defer readerMutex.Unlock()

	if v, _ := k.readers[partitionId]; v != nil {
		_ = v.Close()
	}

	delete(k.readers, partitionId)
}

func (k *KafkaListener) Close() error {
	k.cancelFn()

	runningReq := false

	if k.hasRunningRequest {
		runningReq = true

		for i := 1; i < 5; i++ {
			if !k.hasRunningRequest {
				runningReq = false
				break
			}

			time.Sleep(1 * time.Second)
		}
	}

	for partitionId, _ := range k.readers {
		k.closeReader(partitionId)
	}

	if runningReq {
		return errors.New("kafka listener still has running requests")
	}

	return nil
}

func (k *KafkaListener) listen(maxBatchSize int, maxDuration time.Duration, reader *kafka.Reader) error {
	messagePool := make([]kafka.Message, maxBatchSize)
	messageIndex := 0

	for k.ctx.Err() == nil {
		message2, err := reader.FetchMessage(k.ctx)

		apmTransaction := apm_helper.StartNewApmTransaction(k.listenerName, "kafka_listener", nil,
			nil)

		if err != nil {
			if errors.Is(err, io.EOF) {
				apmTransaction.End()
				return err
			}

			log.Err(err).Send()
			apm_helper.CaptureApmError(err, apmTransaction)
			apmTransaction.End()

			continue
		}

		k.hasRunningRequest = true

		messagePool[0] = message2
		messageIndex = 1

		if maxBatchSize > 1 {
			innerCtx, innerCancelFn := context.WithTimeout(k.ctx, maxDuration)

			for innerCtx.Err() == nil {
				message1, err1 := reader.FetchMessage(innerCtx)

				if err1 == context.DeadlineExceeded {
					break
				}

				if err1 != nil {
					if errors.Is(err1, io.EOF) {
						innerCancelFn()
						k.hasRunningRequest = false

						return err1
					}
					log.Err(err1).Send()
				}

				if err1 == nil {
					messagePool[messageIndex] = message1
					messageIndex += 1
				}

				if messageIndex >= maxBatchSize {
					innerCancelFn()
				}
			}

			innerCancelFn()

			apm_helper.AddApmData(apmTransaction, "messages_count", messageIndex)
		}

		if k.ctx.Err() != nil {
			apmTransaction.Discard()
			k.hasRunningRequest = false
			break // discard messages
		}

		commandExecutionContext := apm.ContextWithTransaction(context.TODO(), apmTransaction)

		_, err = k.command.Execute(structs.ExecutionData{
			ApmTransaction: apmTransaction,
			Context:        commandExecutionContext,
		}, messagePool[:messageIndex]...)

		if err != nil {
			apm_helper.CaptureApmError(err, apmTransaction)
			apmTransaction.End()
			k.hasRunningRequest = false

			return errors.Wrap(err, "re-read")
		}

		if err := reader.CommitMessages(commandExecutionContext, messagePool[:messageIndex]...); err != nil {
			apm_helper.CaptureApmError(err, apmTransaction)

			if errors.Is(err, io.EOF) {
				apmTransaction.End()

				break
			}

			k.hasRunningRequest = false

			apmTransaction.End()
			continue
		}

		k.hasRunningRequest = false
	}

	k.hasRunningRequest = false

	return nil
}
