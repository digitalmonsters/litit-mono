package kafka_listener

import (
	"context"
	"github.com/digitalmonsters/go-common/boilerplate"
	"time"
)

type BatchListener struct {
	innerListener *kafkaListener
	maxDuration   time.Duration
	maxBatchSize  int
}

func NewBatchListener(configuration boilerplate.KafkaListenerConfiguration, command ICommand,
	ctx context.Context, maxDuration time.Duration, maxBatchSize int,
) IKafkaListener {

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

func (b *BatchListener) Listen() {
	b.innerListener.ListenInBatches(b.maxBatchSize, b.maxDuration)
}

func (b *BatchListener) ListenAsync() IKafkaListener {
	go func() {
		b.Listen()
	}()

	return b
}

func (b *BatchListener) Close() error {
	if b.innerListener != nil {
		return b.innerListener.Close()
	}

	return nil
}
