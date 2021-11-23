package kafka_listener

import (
	"context"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/kafka_listener/internal"
	"github.com/digitalmonsters/go-common/kafka_listener/structs"
	"time"
)

type BatchListener struct {
	innerListener *internal.KafkaListener
	maxDuration   time.Duration
	maxBatchSize  int
}

func NewBatchListener(configuration boilerplate.KafkaListenerConfiguration, command structs.ICommand,
	ctx context.Context, maxDuration time.Duration, maxBatchSize int,
) *BatchListener {

	var b = &BatchListener{
		innerListener: internal.NewKafkaListener(configuration, ctx, command),
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

func (b *BatchListener) Close() error {
	if b.innerListener != nil {
		return b.innerListener.Close()
	}

	return nil
}
