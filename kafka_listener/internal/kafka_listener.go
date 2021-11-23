package internal

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/kafka_listener/structs"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
	"go.elastic.co/apm"
	"io"
	"sync"
	"time"
)

var readerMutex sync.Mutex

type KafkaListener struct {
	cfg               boilerplate.KafkaListenerConfiguration
	ctx               context.Context
	reader            *kafka.Reader
	targetTopic       string
	command           structs.ICommand
	listenerName      string
	cancelFn          context.CancelFunc
	hasRunningRequest bool
}

func NewKafkaListener(config boilerplate.KafkaListenerConfiguration, ctx context.Context, command structs.ICommand) *KafkaListener {
	localCtx, cancelFn := context.WithCancel(ctx)

	return &KafkaListener{
		cfg:          config,
		ctx:          localCtx,
		cancelFn:     cancelFn,
		targetTopic:  config.Topic,
		command:      command,
		listenerName: fmt.Sprintf("kafka_listener_%v", config.Topic),
	}
}

func (k KafkaListener) GetTopic() string {
	return k.targetTopic
}

func (k *KafkaListener) connect() (*kafka.Reader, error) {
	if k.reader != nil {
		return k.reader, nil
	}

	readerMutex.Lock()
	defer readerMutex.Unlock()

	if k.reader != nil {
		return k.reader, nil
	}

	var kafkaCfg = kafka.ReaderConfig{
		Brokers:  boilerplate.SplitHostsToSlice(k.cfg.Hosts),
		GroupID:  k.cfg.GroupId,
		Topic:    k.targetTopic,
		MinBytes: k.cfg.MinBytes,
		MaxBytes: k.cfg.MaxBytes,
	}

	if k.cfg.Tls {
		kafkaCfg.Dialer = &kafka.Dialer{
			Timeout:   10 * time.Second,
			DualStack: true,
			TLS: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}

	if k.cfg.KafkaAuth != nil && len(k.cfg.KafkaAuth.Type) > 0 {
		var mech sasl.Mechanism

		if k.cfg.KafkaAuth.Type == "plain" {
			mech = plain.Mechanism{
				Username: "username",
				Password: "password",
			}

		} else {
			mechanism, err := scram.Mechanism(scram.SHA512, k.cfg.KafkaAuth.User, k.cfg.KafkaAuth.Password)

			if err != nil {
				return nil, err
			}
			mech = mechanism
		}
		kafkaCfg.Dialer = &kafka.Dialer{
			Timeout:       10 * time.Second,
			DualStack:     true,
			SASLMechanism: mech,
		}
	}

	r := kafka.NewReader(kafkaCfg)

	k.reader = r
	return r, nil
}

func (k *KafkaListener) ListenInBatches(maxBatchSize int, maxDuration time.Duration) {
	for k.ctx.Err() == nil {
		if err := k.listen(maxBatchSize, maxDuration); err != nil {
			tx := apm_helper.StartNewApmTransaction(k.listenerName, "kafka_listener", nil, nil)

			apm_helper.CaptureApmError(err, tx)
			log.Err(err).Send()

			tx.End()
			time.Sleep(5 * time.Second)
		}
	}

	_ = k.Close()
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

	if k.reader != nil {
		return k.reader.Close()
	}

	if runningReq {
		return errors.New("kafka listener still has running requests")
	}

	return nil
}

func (k *KafkaListener) listen(maxBatchSize int, maxDuration time.Duration) error {
	if _, err := k.connect(); err != nil {
		return errors.WithStack(err)
	}

	messagePool := make([]kafka.Message, maxBatchSize)
	messageIndex := 0

	offset := int64(0)

	for k.ctx.Err() == nil {
		message2, err := k.reader.FetchMessage(k.ctx)

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
		offset = message2.Offset

		if maxBatchSize > 1 {
			innerCtx, innerCancelFn := context.WithTimeout(k.ctx, maxDuration)

			for innerCtx.Err() == nil {
				message1, err1 := k.reader.FetchMessage(innerCtx)

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

			if err = k.reader.SetOffset(offset); err != nil {
				apm_helper.CaptureApmError(err, apmTransaction)
			}

			apmTransaction.End()
			k.hasRunningRequest = false

			continue
		}

		if err := k.reader.CommitMessages(commandExecutionContext, messagePool[:messageIndex]...); err != nil {
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
