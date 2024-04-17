package eventsourcing

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"time"

	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
	"go.elastic.co/apm/module/apmhttp"
)

type KafkaEventPublisher struct {
	writer        *kafka.Writer
	cfg           boilerplate.KafkaWriterConfiguration
	topic         string
	publisherType PublisherType
	logger        zerolog.Logger
	firstHost     string
}

func NewKafkaEventPublisher(cfg boilerplate.KafkaWriterConfiguration, topicConfig boilerplate.KafkaTopicConfig) *KafkaEventPublisher {
	hosts := boilerplate.SplitHostsToSlice(cfg.Hosts)

	h := &KafkaEventPublisher{
		cfg: cfg,
		writer: &kafka.Writer{
			Addr:         kafka.TCP(hosts...),
			Topic:        topicConfig.Name,
			Balancer:     &kafka.Hash{},
			BatchTimeout: 10 * time.Millisecond,
		},
		topic:         topicConfig.Name,
		publisherType: PublisherTypeKafka,
		logger:        log.Logger.With().Str("topic", topicConfig.Name).Logger(),
		firstHost:     hosts[0],
	}

	if cfg.Tls {
		dialer := kafka.DefaultDialer
		dialer.TLS = &tls.Config{
			InsecureSkipVerify: true,
		}

		h.writer.Transport = &kafka.Transport{
			TLS: &tls.Config{
				InsecureSkipVerify: true,
			},
			Dial: dialer.DialFunc,
		}
	}

	h.ensureTopicExists(topicConfig)

	return h
}

func (s *KafkaEventPublisher) Publish(apmTransaction *apm.Transaction, events ...IEventData) []error {
	if len(events) == 0 {
		return nil
	}

	var sp *apm.Span

	if apmTransaction != nil {
		serviceName := fmt.Sprintf("%v [%v]", s.firstHost, s.topic)

		sp = apmTransaction.StartSpan(fmt.Sprintf("kafka publish [%v]", serviceName), "kafka", nil)
		sp.Context.SetLabel("count", len(events))

		sp.Context.SetMessage(apm.MessageSpanContext{
			QueueName: serviceName,
		})

		sp.Context.SetDatabaseRowsAffected(int64(len(events)))
		sp.Context.SetDestinationService(apm.DestinationServiceSpanContext{
			Name:     "kafka",
			Resource: serviceName,
		})

		defer func() {
			sp.End()
		}()
	}

	var eventsMarshalled []kafka.Message

	for _, event := range events {
		value, err := json.Marshal(event)

		if err != nil {
			return []error{errors.WithStack(err)}
		}

		var headers []kafka.Header

		if apmTransaction != nil {
			headers = append(headers, kafka.Header{
				Key:   apmhttp.W3CTraceparentHeader,
				Value: []byte(apmhttp.FormatTraceparentHeader(apmTransaction.TraceContext())),
			})
		}

		eventsMarshalled = append(eventsMarshalled, kafka.Message{
			Key:     []byte(event.GetPublishKey()),
			Value:   value,
			Time:    time.Now().UTC(),
			Headers: headers,
		})
	}

	if err := s.writer.WriteMessages(context.Background(), eventsMarshalled...); err != nil {
		return []error{errors.WithStack(err)}
	}

	return nil
}

func (s *KafkaEventPublisher) GetPublisherType() PublisherType {
	return s.publisherType
}

func (s *KafkaEventPublisher) GetHosts() string {
	return s.cfg.Hosts
}

func (s *KafkaEventPublisher) GetTopic() string {
	return s.topic
}

func (s *KafkaEventPublisher) ensureTopicExists(topicConfig boilerplate.KafkaTopicConfig) {
	client := &kafka.Client{
		Transport: s.writer.Transport,
	}
	meta, err := client.Metadata(context.TODO(), &kafka.MetadataRequest{
		Addr: s.writer.Addr,
	})

	if err != nil {
		s.logger.Fatal().Err(err).Str("host", s.cfg.Hosts).Str("topic", s.topic).Msg("[Kafka Publisher] : can not ensure that topic exists")
	}

	var exists bool
	for _, t := range meta.Topics {
		if t.Name == topicConfig.Name {
			exists = true
			if len(t.Partitions) != topicConfig.NumPartitions {
				s.logger.Warn().Str("host", s.cfg.Hosts).Str("topic", s.topic).Msgf("[Kafka Publisher] : partition count mismatch for topic [%v] expected [%v] got [%v] partitions",
					topicConfig.Name, topicConfig.NumPartitions, len(t.Partitions))
			}
			break
		}
	}

	if !exists {
		if topicConfig.NumPartitions <= 0 {
			s.logger.Panic().Str("host", s.cfg.Hosts).Str("topic", s.topic).Msg("[Kafka Publisher] : NumPartitions should be greater then 0")
		}
		if topicConfig.ReplicationFactor <= 0 {
			s.logger.Panic().Str("host", s.cfg.Hosts).Str("topic", s.topic).Msg("[Kafka Publisher] : ReplicationFactor should be greater then 0")
		}

		s.logger.Info().Str("host", s.cfg.Hosts).Str("topic", s.topic).Int("partition", topicConfig.NumPartitions).
			Int("replicationFactor", topicConfig.ReplicationFactor).Msg("[Kafka Publisher] : topic does not exist on kafka. Creating a new one.")

		res, err := client.CreateTopics(context.TODO(), &kafka.CreateTopicsRequest{
			Addr: s.writer.Addr,
			Topics: []kafka.TopicConfig{
				{
					Topic:             topicConfig.Name,
					NumPartitions:     topicConfig.NumPartitions,
					ReplicationFactor: topicConfig.ReplicationFactor,
				},
			},
			ValidateOnly: false,
		})

		if err != nil {
			s.logger.Fatal().Err(err).Str("host", s.cfg.Hosts).Str("topic", s.topic).Msg("[Kafka Publisher] : can not create topic")
		}

		for _, respErr := range res.Errors {
			if respErr != nil {
				s.logger.Fatal().Err(respErr).Str("host", s.cfg.Hosts).Str("topic", s.topic).Msg("[Kafka Publisher] : create topic failed")
			}
		}
	}
}
