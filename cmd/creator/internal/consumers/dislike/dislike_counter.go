package dislike

import (
	"context"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/kafka_listener"
	"github.com/digitalmonsters/music/configs"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/rs/zerolog"

	"github.com/segmentio/kafka-go"
	"time"
)

type dislikeCounter struct {
	name           string
	listener       kafka_listener.IKafkaListener
	logger         zerolog.Logger
	kafkaConfig    configs.CounterListener
	applicationCtx context.Context
}

func SubApp(applicationCtx context.Context, kafkaConfig configs.CounterListener) application.SubApplication {
	return &dislikeCounter{
		kafkaConfig:    kafkaConfig,
		name:           "dislike_counter",
		applicationCtx: applicationCtx,
	}
}

func (l *dislikeCounter) Init(subAppLogger zerolog.Logger) error {
	l.logger = subAppLogger
	service := newLikeCounterService(l.kafkaConfig.WorkerPoolSize)

	l.listener = kafka_listener.NewBatchListener(l.kafkaConfig.Kafka, kafka_listener.NewCommand("dislike_counter_service",
		func(executionData kafka_listener.ExecutionData, request ...kafka.Message) []kafka.Message {
			mapped, mapErrors, messagesToCommit := mapKafkaMessages(request)
			for _, err := range mapErrors {
				apm_helper.LogError(err, executionData.Context)
			}

			if len(mapped) < 10 {
				apm_helper.AddApmData(executionData.ApmTransaction, "data", mapped)
			}

			messagesToCommit = append(messagesToCommit, service.Process(mapped,
				database.GetDbWithContext(database.DbTypeMaster, executionData.Context),
				executionData.ApmTransaction, executionData.Context)...)

			return messagesToCommit
		}, false), l.applicationCtx, time.Duration(l.kafkaConfig.MaxDuration)*time.Millisecond, l.kafkaConfig.MaxBatchSize).ListenAsync()

	return nil
}

func (l dislikeCounter) Name() string {
	return l.name
}

func (l dislikeCounter) Close() error {
	return l.listener.Close()
}
