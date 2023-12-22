package kafka_listener

import (
	"context"

	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
)

type IKafkaListener interface {
	Close() error
	Listen(createTopicIfNotFound bool)
	ListenAsync(createTopicIfNotFound ...bool) IKafkaListener
	GetTopic() string
	GetHosts() string
}

type ICommand interface {
	Execute(executionData ExecutionData, request ...kafka.Message) (successfullyProcessed []kafka.Message)
	GetFancyName() string
	ForceLog() bool
}

type ExecutionData struct {
	ApmTransaction *apm.Transaction
	Context        context.Context
}

type ErrorWithKafkaMessage struct {
	Error   error
	Message kafka.Message
}
