package kafka_listener

import (
	"context"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/kafka_listener/internal"
	"github.com/digitalmonsters/go-common/kafka_listener/structs"
)

type SingleListener struct {
	listener *internal.KafkaListener
}

func NewSingleListener(configuration boilerplate.KafkaListenerConfiguration, command structs.ICommand,
	ctx context.Context) *SingleListener {

	var s = &SingleListener{
		listener: internal.NewKafkaListener(configuration, ctx, command),
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
