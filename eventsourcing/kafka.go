package eventsourcing

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
	"time"
)

type KafkaEventPublisher struct {
	writer        *kafka.Writer
	cfg           boilerplate.KafkaWriterConfiguration
	topic         string
	publisherType PublisherType
	logger        zerolog.Logger
}

func NewKafkaEventPublisher(cfg boilerplate.KafkaWriterConfiguration, topicConfig boilerplate.KafkaTopicConfig) *KafkaEventPublisher {
	h := &KafkaEventPublisher{
		cfg: cfg,
		writer: &kafka.Writer{
			Addr:         kafka.TCP(boilerplate.SplitHostsToSlice(cfg.Hosts)...),
			Topic:        topicConfig.Name,
			Balancer:     &kafka.Hash{},
			BatchTimeout: 10 * time.Millisecond,
		},
		topic:         topicConfig.Name,
		publisherType: PublisherTypeKafka,
		logger:        log.Logger.With().Str("topic", topicConfig.Name).Logger(),
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
		sp = apmTransaction.StartSpan(fmt.Sprintf("kafka publish [%v]", s.topic), "kafka", nil)
		sp.Context.SetLabel("count", len(events))

		sp.Context.SetMessage(apm.MessageSpanContext{
			QueueName: s.topic,
		})
		sp.Context.SetDatabaseRowsAffected(int64(len(events)))
		sp.Context.SetDestinationService(apm.DestinationServiceSpanContext{
			Name:     "kafka",
			Resource: s.topic,
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

		eventsMarshalled = append(eventsMarshalled, kafka.Message{
			Key:   []byte(event.GetPublishKey()),
			Value: value,
			Time:  time.Now().UTC(),
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

func (s *KafkaEventPublisher) ensureTopicExists(topicConfig boilerplate.KafkaTopicConfig) {
	client := &kafka.Client{
		Transport: s.writer.Transport,
	}
	meta, err := client.Metadata(context.TODO(), &kafka.MetadataRequest{
		Addr: s.writer.Addr,
	})

	if err != nil {
		s.logger.Fatal().Err(err).Msgf("can not ensure that topic exists [%v]", topicConfig.Name)
	}

	var exists bool
	for _, t := range meta.Topics {
		if t.Name == topicConfig.Name {
			exists = true
			if len(t.Partitions) != topicConfig.NumPartitions {
				s.logger.Warn().Msgf("partition count mismatch for topic [%v] expected [%v] got [%v] partitions",
					topicConfig.Name, topicConfig.NumPartitions, t.Partitions)
			}
			break
		}
	}

	if !exists {
		if topicConfig.NumPartitions <= 0 {
			s.logger.Panic().Msgf("NumPartitions should be greater then 0 for topic [%v]", topicConfig.Name)
		}
		if topicConfig.ReplicationFactor <= 0 {
			s.logger.Panic().Msgf("ReplicationFactor should be greater then 0 for topic [%v]", topicConfig.Name)
		}

		s.logger.Info().Msgf("topic [%v] does not exist on kafka. Creating a new one with partitions count [%v] and replication factor [%v]",
			topicConfig.Name, topicConfig.NumPartitions, topicConfig.ReplicationFactor)

		res, err := client.CreateTopics(context.TODO(), &kafka.CreateTopicsRequest{
			Addr: s.writer.Addr,
			Topics: []kafka.TopicConfig{
				kafka.TopicConfig{
					Topic:             topicConfig.Name,
					NumPartitions:     topicConfig.NumPartitions,
					ReplicationFactor: topicConfig.ReplicationFactor,
				},
			},
			ValidateOnly: false,
		})

		if err != nil {
			s.logger.Fatal().Err(err).Msgf("can not create topic [%v]", topicConfig.Name)
		}

		if len(res.Errors) > 0 {
			for _, respErr := range res.Errors {
				s.logger.Fatal().Err(respErr).Msgf("can not create topic [%v]", topicConfig.Name)
			}
		}
	}
}
