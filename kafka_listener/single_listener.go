package kafka_listener

import (
	"context"
	"github.com/digitalmonsters/go-common/boilerplate"
)

type SingleListener struct {
	listener *kafkaListener
}

func NewSingleListener(configuration boilerplate.KafkaListenerConfiguration, command ICommand,
	ctx context.Context) IKafkaListener {

	var s = &SingleListener{
		listener: newKafkaListener(configuration, ctx, command),
	}

	return s
}

func (s *SingleListener) Close() error {
	if s.listener != nil {
		return s.listener.Close()
	}

	return nil
}

func (s *SingleListener) Listen() {
	s.listener.ListenInBatches(1, 0)
}

func (s *SingleListener) ListenAsync() IKafkaListener {
	go func() {
		s.Listen()
	}()

	return s
}

func (s SingleListener) GetTopic() string {
	return s.listener.GetTopic()
}
