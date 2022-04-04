package user_delete

import (
	"context"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/kafka_listener"
	"github.com/segmentio/kafka-go"
)

func InitListener(appCtx context.Context, configuration boilerplate.KafkaListenerConfiguration) kafka_listener.IKafkaListener {
	return kafka_listener.NewSingleListener(configuration, kafka_listener.NewCommand("user delete",
		func(executionData kafka_listener.ExecutionData, request ...kafka.Message) []kafka.Message {
			singleMessage := request[0]

			// nolint
			if 1 == 1 {
				return request // temp mock
			}
			
			mapped, err := mapKafkaMessages(singleMessage)

			if err != nil {
				apm_helper.CaptureApmError(err, executionData.ApmTransaction)

				return []kafka.Message{singleMessage}
			}

			result, err := process(*mapped, executionData.Context)

			if err != nil {
				apm_helper.CaptureApmError(err, executionData.ApmTransaction)
			}

			successfulMessages := make([]kafka.Message, 0)

			if result != nil {
				successfulMessages = append(successfulMessages, *result)
			}

			return successfulMessages
		}, true), appCtx)
}
