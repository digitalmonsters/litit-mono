package eventsourcing

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
)

type KafkaEventPublisher struct {
	writer        *kafka.Writer
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

	return h
}

func (s *KafkaEventPublisher) Publish(apmTransaction *apm.Transaction, events ...IEventData) []error {
	
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
