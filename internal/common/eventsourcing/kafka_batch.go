package eventsourcing

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
	"go.elastic.co/apm/module/apmhttp"
)

type kafkaRecord struct {
	message      kafka.Message
	maxRetry     int
	currentRetry int
	traceContext string
}

type iMessageWriter interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
}

type mockWriter struct {
	WriteFn func(ctx context.Context, msgs ...kafka.Message) error
}

func (m *mockWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	return m.WriteFn(ctx, msgs...)
}

type ciRun struct {
}

type KafkaEventPublisherV2[T IEventData] struct {
	name               string
	cfg                boilerplate.KafkaBatchWriterV2Configuration
	topicConfig        boilerplate.KafkaTopicConfig
	queue              []kafkaRecord
	mut                sync.Mutex
	flushMut           sync.Mutex
	ctx                context.Context
	writer             iMessageWriter
	logger             zerolog.Logger
	messagesIncoming   prometheus.Counter
	messagesDropped    prometheus.Counter
	isClosed           bool
	fancyServiceMapLog bool
	firstHost          string
}

type Publisher[T IEventData] interface {
	Publish(ctx context.Context, messages ...T) chan error
	PublishImmediate(ctx context.Context, messages ...T) chan error
	GetHosts() string
	GetTopic() string
	Close() error
}

type PublisherMock[T IEventData] struct {
	PublishFn          func(ctx context.Context, messages ...T) chan error
	PublishImmediateFn func(ctx context.Context, messages ...T) chan error
	GetHostsFn         func() string
	GetTopicFn         func() string
	CloseFn            func() error
}

func (p *PublisherMock[T]) PublishImmediate(ctx context.Context, messages ...T) chan error {
	return p.PublishImmediateFn(ctx, messages...)

}

func (p *PublisherMock[T]) GetHosts() string {
	return p.GetHostsFn()
}

func (p *PublisherMock[T]) GetTopic() string {
	return p.GetTopicFn()
}

func (p *PublisherMock[T]) Close() error {
	return p.CloseFn()
}

func (p *PublisherMock[T]) Publish(ctx context.Context, messages ...T) chan error {
	return p.PublishFn(ctx, messages...)
}

func NewMock[T IEventData]() Publisher[T] {
	return &PublisherMock[T]{}
}

var registrationMap = map[string]bool{}
var registrationMut = sync.Mutex{}

func NewKafkaBatchPublisher[T IEventData](publisherName string, cfg boilerplate.KafkaBatchWriterV2Configuration,
	ctx context.Context) Publisher[T] {
	hosts := boilerplate.SplitHostsToSlice(cfg.Hosts)

	registrationMut.Lock()

	if _, ok := registrationMap[publisherName]; ok {
		panic(fmt.Sprintf("kafka publisher with name [%v] already registered", publisherName))
	}

	registrationMap[publisherName] = true
	registrationMut.Unlock()

	writer := &kafka.Writer{
		Addr:         kafka.TCP(hosts...),
		Topic:        cfg.Topic.Name,
		Balancer:     &kafka.Hash{},
		BatchTimeout: 50 * time.Millisecond,
	}

	if len(hosts) == 0 { // test only
		hosts = append(hosts, "unk")
	}

	p := &KafkaEventPublisherV2[T]{
		name:   publisherName,
		writer: writer,
		queue:  make([]kafkaRecord, 0),
		logger: log.Logger.With().Str("topic", cfg.Topic.Name).
			Str("publisher_name", publisherName).
			Logger(),
		ctx: ctx,
		messagesIncoming: promauto.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("messages_incoming_%v", publisherName),
			Help: "Number of incoming messages",
		}),
		messagesDropped: promauto.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("messages_dropped_%v", publisherName),
			Help: "Number of dropped messages",
		}),
		firstHost:          hosts[0],
		fancyServiceMapLog: true,
	}

	if cfg.FlushTimeMilliseconds == 0 {
		cfg.FlushTimeMilliseconds = 100
	}

	cfg.FlushTimeMilliseconds = cfg.FlushTimeMilliseconds * int(time.Millisecond)

	if cfg.BackOffTimeMaxMilliseconds <= 0 {
		cfg.BackOffTimeMaxMilliseconds = 30 * 1000 // 30 sec
	}

	cfg.BackOffTimeMaxMilliseconds = cfg.BackOffTimeMaxMilliseconds * int(time.Millisecond)

	if cfg.BackOffTimeIntervalMilliseconds <= 0 {
		cfg.BackOffTimeIntervalMilliseconds = 1 * 1000 // 1 sec
	}

	cfg.BackOffTimeIntervalMilliseconds = cfg.BackOffTimeIntervalMilliseconds * int(time.Millisecond)

	if cfg.FlushAtSize == 0 {
		cfg.FlushAtSize = 20
	}

	if cfg.MaxRetryCount <= 0 {
		cfg.MaxRetryCount = 3
	}

	if cfg.Tls {
		dialer := kafka.DefaultDialer
		dialer.TLS = &tls.Config{
			InsecureSkipVerify: true,
		}

		writer.Transport = &kafka.Transport{
			TLS: &tls.Config{
				InsecureSkipVerify: true,
			},
			Dial: dialer.DialFunc,
		}
	}

	p.cfg = cfg
	p.topicConfig = cfg.Topic

	p.startAsync()

	if v := ctx.Value(ciRun{}); v != nil { // test run
		return p
	}

	p.ensureTopicExists(cfg.Topic)

	return p
}

func (p *KafkaEventPublisherV2[T]) Publish(ctx context.Context, messages ...T) chan error {
	ch := make(chan error, 2)

	go func() {
		defer func() {
			close(ch)
		}()

		if p.isClosed {
			ch <- errors.New("publisher is closed already")

			return
		}

		msgs, err := p.prepareMessages(ctx, messages...)
		if err != nil {
			ch <- err
			return
		}

		p.addToQueue(true, msgs...)
	}()

	return ch
}

func (p *KafkaEventPublisherV2[T]) PublishImmediate(ctx context.Context, messages ...T) chan error {
	ch := make(chan error, 2)

	go func() {
		defer func() {
			close(ch)
		}()

		if p.isClosed {
			ch <- errors.New("publisher is closed already")

			return
		}

		msg, err := p.prepareMessages(ctx, messages...)
		if err != nil {
			ch <- err
			return
		}

		p.addToQueue(false, msg...)

		if err = p.flush(true); err != nil {
			ch <- err
			return
		}

		ch <- nil
	}()

	return ch
}

func (p *KafkaEventPublisherV2[T]) GetHosts() string {
	return p.cfg.Hosts
}

func (p *KafkaEventPublisherV2[T]) GetTopic() string {
	return p.topicConfig.Name
}

func (p *KafkaEventPublisherV2[T]) Close() error {
	p.isClosed = true

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 8 * time.Second
	b.InitialInterval = 100 * time.Millisecond
	b.Reset()

	return backoff.Retry(func() error {
		return p.flush(false)
	}, b)
}

func (p *KafkaEventPublisherV2[T]) sendBatch(msg ...kafkaRecord) error {
	kafkaMgs := make([]kafka.Message, len(msg))

	for i, m := range msg {
		kafkaMgs[i] = m.message
	}

	return p.writer.WriteMessages(p.ctx, kafkaMgs...)
}

func (p *KafkaEventPublisherV2[T]) flush(calculateAsRetry bool) error {
	p.flushMut.Lock()
	defer p.flushMut.Unlock()

	p.mut.Lock()
	if len(p.queue) == 0 {
		p.mut.Unlock()
		return nil
	}

	batch := p.queue
	p.queue = make([]kafkaRecord, 0)
	p.mut.Unlock()

	var fancyApmTx *apm.Transaction

	if p.fancyServiceMapLog { // fancy logging for APM ServiceMap
		p.fancyServiceMapLog = false

		fancyApmTx = apm_helper.StartNewApmTransaction(p.name, "publisher",
			nil, nil)

		serviceName := p.topicConfig.Name

		sp := fancyApmTx.StartSpan(fmt.Sprintf("kafka publish [%v]", serviceName), "kafka", nil)
		sp.Context.SetLabel("count", len(batch))

		sp.Context.SetMessage(apm.MessageSpanContext{
			QueueName: serviceName,
		})

		sp.Context.SetDatabaseRowsAffected(int64(len(batch)))
		sp.Context.SetDestinationService(apm.DestinationServiceSpanContext{
			Name:     p.firstHost,
			Resource: p.topicConfig.Name,
		})

		defer func() {
			sp.End()
			fancyApmTx.End()
		}()
	}

	err := p.sendBatch(batch...)

	if err == nil {
		return nil // we are good
	}

	// lets try to requeue
	toEnqueue := make([]kafkaRecord, 0)
	var droppedMessages []kafkaRecord

	for _, b := range batch {
		if calculateAsRetry {
			b.currentRetry += 1

			if b.currentRetry > b.maxRetry {
				droppedMessages = append(droppedMessages, b)
				continue
			}
		}

		toEnqueue = append(toEnqueue, b)
	}

	if len(droppedMessages) > 0 {
		p.messagesDropped.Add(float64(len(droppedMessages)))

		if fancyApmTx == nil {
			fancyApmTx = apm_helper.StartNewApmTransaction(p.name, "publisher",
				nil, nil)

			defer func() {
				fancyApmTx.End()
			}()
		}

		ctx := boilerplate.CreateCustomContext(context.TODO(), fancyApmTx, p.logger)

		apm_helper.AddApmData(fancyApmTx, "dropped", droppedMessages)

		apm_helper.LogError(errors.New("some messages are dropped due to problems with publishing"), ctx)
	}

	if len(toEnqueue) > 0 {
		p.mut.Lock()
		p.queue = append(toEnqueue, p.queue...) // put to head
		p.mut.Unlock()
	}

	return err
}

// nolint
func (p *KafkaEventPublisherV2[T]) fancyServiceMapAsync() {
	go func() {
		for !p.isClosed {
			time.Sleep(3 * time.Minute)

			p.fancyServiceMapLog = true
		}
	}()
}
func (p *KafkaEventPublisherV2[T]) startAsync() {
	go func() {
		for !p.isClosed {
			time.Sleep(time.Duration(p.cfg.FlushTimeMilliseconds))

			b := backoff.NewExponentialBackOff()
			b.MaxElapsedTime = time.Duration(p.cfg.BackOffTimeMaxMilliseconds)
			b.InitialInterval = time.Duration(p.cfg.BackOffTimeIntervalMilliseconds)
			b.Reset()

			if err := backoff.Retry(func() error {
				return p.flush(true)
			}, b); err != nil {
				apmTransaction := apm_helper.StartNewApmTransaction(p.name, "publisher",
					nil, nil)

				ctx := boilerplate.CreateCustomContext(context.TODO(), apmTransaction, p.logger)

				apm_helper.LogError(err, ctx)

				apmTransaction.End()
			}
		}
	}()
}

func (p *KafkaEventPublisherV2[T]) addToQueue(triggerFlushAtSize bool, messages ...kafkaRecord) {
	p.mut.Lock()
	p.queue = append(p.queue, messages...)
	p.mut.Unlock()

	if triggerFlushAtSize && len(p.queue) >= p.cfg.FlushAtSize {
		if err := p.flush(false); err != nil {
			apmTransaction := apm_helper.StartNewApmTransaction(p.name, "publisher",
				nil, nil)

			ctx := boilerplate.CreateCustomContext(context.TODO(), apmTransaction, p.logger)

			apm_helper.LogError(err, ctx)

			apmTransaction.End()
		}
	}
}

func (p *KafkaEventPublisherV2[T]) prepareMessages(ctx context.Context, messages ...T) ([]kafkaRecord, error) {
	toSend := make([]kafkaRecord, len(messages))

	p.messagesIncoming.Add(float64(len(messages)))

	for i, m := range messages {
		value, err := json.Marshal(m)

		if err != nil {
			p.messagesDropped.Add(float64(len(messages)))
			return nil, errors.WithStack(err)
		}

		var headers []kafka.Header

		var traceContext string
		if apmTransaction := apm.TransactionFromContext(ctx); apmTransaction != nil {
			traceContext = apmhttp.FormatTraceparentHeader(apmTransaction.TraceContext())

			headers = append(headers, kafka.Header{
				Key:   apmhttp.W3CTraceparentHeader,
				Value: []byte(traceContext),
			})
		}

		toSend[i] = kafkaRecord{
			message: kafka.Message{
				Key:     []byte(m.GetPublishKey()),
				Value:   value,
				Time:    time.Now().UTC(),
				Headers: headers,
			},
			maxRetry:     p.cfg.MaxRetryCount,
			traceContext: traceContext,
		}
	}

	return toSend, nil
}

func (p *KafkaEventPublisherV2[T]) ensureTopicExists(topicConfig boilerplate.KafkaTopicConfig) {
	writer := p.writer.(*kafka.Writer)
	client := &kafka.Client{
		Transport: writer.Transport,
	}
	meta, err := client.Metadata(context.TODO(), &kafka.MetadataRequest{
		Addr: writer.Addr,
	})

	if err != nil {
		p.logger.Fatal().Err(err).Msgf("can not ensure that topic exists [%v]", topicConfig.Name)
	}

	var exists bool
	for _, t := range meta.Topics {
		if t.Name == topicConfig.Name {
			exists = true
			if len(t.Partitions) != topicConfig.NumPartitions {
				p.logger.Warn().Msgf("partition count mismatch for topic [%v] expected [%v] got [%v] partitions",
					topicConfig.Name, topicConfig.NumPartitions, len(t.Partitions))
			}
			break
		}
	}

	if !exists {
		if topicConfig.NumPartitions <= 0 {
			p.logger.Panic().Msgf("NumPartitions should be greater then 0 for topic [%v]", topicConfig.Name)
		}
		if topicConfig.ReplicationFactor <= 0 {
			p.logger.Panic().Msgf("ReplicationFactor should be greater then 0 for topic [%v]", topicConfig.Name)
		}

		p.logger.Info().Msgf("topic [%v] does not exist on kafka. Creating a new one with partitions count [%v] and replication factor [%v]",
			topicConfig.Name, topicConfig.NumPartitions, topicConfig.ReplicationFactor)

		realTopicConfig := kafka.TopicConfig{
			Topic:             topicConfig.Name,
			NumPartitions:     topicConfig.NumPartitions,
			ReplicationFactor: topicConfig.ReplicationFactor,
		}

		if topicConfig.RetentionMs > 0 {
			realTopicConfig.ConfigEntries = append(realTopicConfig.ConfigEntries,
				kafka.ConfigEntry{
					ConfigName:  "retention.ms",
					ConfigValue: fmt.Sprint(topicConfig.RetentionMs),
				})
		}

		res, err := client.CreateTopics(context.TODO(), &kafka.CreateTopicsRequest{
			Addr: writer.Addr,
			Topics: []kafka.TopicConfig{
				realTopicConfig,
			},
			ValidateOnly: false,
		})

		if err != nil {
			p.logger.Fatal().Err(err).Msgf("can not create topic [%v]", topicConfig.Name)
		}

		for _, respErr := range res.Errors {
			if respErr != nil {
				p.logger.Fatal().Err(respErr).Msgf("can not create topic [%v]", topicConfig.Name)
			}
		}
	}
}
