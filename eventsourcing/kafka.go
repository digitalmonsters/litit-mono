package eventsourcing

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
	"sync"
	"time"
)

type KafkaEventPublisher struct {
	writer        *kafka.Writer
	initMutex     sync.Mutex
	isInitialized bool
	cfg           boilerplate.KafkaWriterConfiguration
	publisherType PublisherType
}

func NewKafkaEventPublisher(cfg boilerplate.KafkaWriterConfiguration, topic string) *KafkaEventPublisher {
	h := &KafkaEventPublisher{
		cfg: cfg,
		writer: &kafka.Writer{
			Addr:     kafka.TCP(boilerplate.SplitHostsToSlice(cfg.Hosts)...),
			Topic:    topic,
			Balancer: &kafka.Hash{},
		},
		publisherType: PublisherTypeKafka,
	}

	if cfg.Tls {
		h.writer.Transport = &kafka.Transport{
			TLS: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}

	return h
}

func (s *KafkaEventPublisher) init() error {
	if s.isInitialized {
		return nil
	}

	s.initMutex.Lock()
	defer s.initMutex.Unlock()

	if s.isInitialized {
		return nil
	}

	dialer := kafka.DefaultDialer

	if s.cfg.Tls {
		dialer = &kafka.Dialer{
			Timeout:   10 * time.Second,
			DualStack: true,
			TLS: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}

	conn, err := dialer.Dial("tcp", s.writer.Addr.String())

	if err != nil {
		return errors.WithStack(err)
	}

	defer func() {
		_ = conn.Close()
	}()

	//if err := conn.CreateTopics(kafka.TopicConfig{
	//	Topic:             s.writer.Topic,
	//	NumPartitions:     s.cfg.TopicPartitionCount,
	//	ReplicationFactor: s.cfg.TopicReplicationFactor,
	//}); err != nil {
	//	return errors.WithStack(err)
	//}

	s.isInitialized = true

	return nil
}

func (s *KafkaEventPublisher) Publish(apmTransaction *apm.Transaction, events ...IEventData) []error {
	if err := s.init(); err != nil {
		return []error{errors.WithStack(err)}
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
