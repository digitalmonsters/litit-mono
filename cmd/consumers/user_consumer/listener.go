package user_consumer

import (
	"context"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/kafka_listener"
	"github.com/segmentio/kafka-go"
)

func InitListener(applicationCtx context.Context, kafkaConfig boilerplate.KafkaListenerConfiguration) kafka_listener.IKafkaListener {

	return kafka_listener.NewSingleListener(kafkaConfig, kafka_listener.NewCommand("user event listener",
		func(executionData kafka_listener.ExecutionData, request ...kafka.Message) []kafka.Message {
			singleMessage := request[0]

			mapped, err := mapKafkaMessages(singleMessage)

			if err != nil {
				apm_helper.CaptureApmError(err, executionData.ApmTransaction)

				return []kafka.Message{singleMessage}
			}

			result, err := process(*mapped, executionData.ApmTransaction, executionData.Context)

			if err != nil {
				apm_helper.CaptureApmError(err, executionData.ApmTransaction)
			}

			successfulMessages := make([]kafka.Message, 0)

			if result != nil {
				successfulMessages = append(successfulMessages, *result)
			}

			return successfulMessages
		}, true), applicationCtx)
}
