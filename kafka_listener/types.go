package kafka_listener

import (
	"context"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
)

type IKafkaListener interface {
	Close() error
	Listen()
	ListenAsync() IKafkaListener
	GetTopic() string
}

type ICommand interface {
	Execute(executionData ExecutionData, request ...kafka.Message) (successfullyProcessed []kafka.Message)
	GetFancyName() string
}

type ExecutionData struct {
	ApmTransaction *apm.Transaction
	Context        context.Context
}
