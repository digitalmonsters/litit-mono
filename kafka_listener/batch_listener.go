package kafka_listener

import (
	"context"
	"time"

	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/rs/zerolog/log"
)

type BatchListener struct {
	innerListener *kafkaListener
	maxDuration   time.Duration
	maxBatchSize  int
}

func NewBatchListener(configuration boilerplate.KafkaListenerConfiguration, command ICommand,
	ctx context.Context, maxDuration time.Duration, maxBatchSize int,
) IKafkaListener {

	if maxBatchSize == 0 {
		maxBatchSize = 1
		log.Warn().Msgf("max batch size is invalid for [%v] settings 1 as max batch", configuration.Topic)
	}

	var b = &BatchListener{
		innerListener: newKafkaListener(configuration, ctx, command),
		maxDuration:   maxDuration,
		maxBatchSize:  maxBatchSize,
	}

	return b
}

func (b BatchListener) GetTopic() string {
	return b.innerListener.GetTopic()
}

func (b BatchListener) GetHosts() string {
	return b.innerListener.GetHosts()
}

func (b *BatchListener) Listen(createTopicIfNotFound bool) {
	b.innerListener.ListenInBatches(b.maxBatchSize, b.maxDuration, createTopicIfNotFound)
}

func (b *BatchListener) ListenAsync(createTopicIfNotFound ...bool) IKafkaListener {
	if len(createTopicIfNotFound) > 1 {
		panic("createTopicIfNotFound can be only one value")
	}

	if len(createTopicIfNotFound) == 0 {
		createTopicIfNotFound = []bool{boilerplate.InLocal()}
	}

	go func() {
		b.Listen(createTopicIfNotFound[0])
	}()

	return b
}

func (b *BatchListener) Close() error {
	if b.innerListener != nil {
		return b.innerListener.Close()
	}

	return nil
}
