package comments_counter

import (
	"context"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/kafka_listener"
	"github.com/rs/zerolog"

	"github.com/segmentio/kafka-go"
)

type commentsCounter struct {
	name           string
	listener       kafka_listener.IKafkaListener
	logger         zerolog.Logger
	kafkaConfig    boilerplate.KafkaListenerConfiguration
	applicationCtx context.Context
}

func SubApp(applicationCtx context.Context, kafkaConfig boilerplate.KafkaListenerConfiguration) application.SubApplication {
	return &commentsCounter{
		kafkaConfig:    kafkaConfig,
		name:           "comments_counter",
		applicationCtx: applicationCtx,
	}
}

func (l *commentsCounter) Init(subAppLogger zerolog.Logger) error {
	l.logger = subAppLogger

	l.listener = kafka_listener.NewSingleListener(l.kafkaConfig, kafka_listener.NewCommand("comments_counter_service",
		func(executionData kafka_listener.ExecutionData, request ...kafka.Message) []kafka.Message {
			singleMessage := request[0]

			mapped, err := mapKafkaMessages(singleMessage)
			if err != nil {
				apm_helper.LogError(err, executionData.Context)
				return []kafka.Message{singleMessage}
			}

			result, err := process(*mapped, executionData)
			if err != nil {
				apm_helper.LogError(err, executionData.Context)
			}

			successfulMessages := make([]kafka.Message, 0)
			if result != nil {
				successfulMessages = append(successfulMessages, *result)
			}

			return successfulMessages
		}, false), l.applicationCtx).ListenAsync()

	return nil
}

func (l commentsCounter) Name() string {
	return l.name
}

func (l commentsCounter) Close() error {
	return l.listener.Close()
}
