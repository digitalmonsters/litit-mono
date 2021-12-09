package eventsourcing

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
	"time"
)

type KafkaEventPublisher struct {
	writer        *kafka.Writer
	cfg           boilerplate.KafkaWriterConfiguration
	topic         string
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
		topic:         topic,
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
