package tokenomics_notification

import (
	"context"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/kafka_listener"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/segmentio/kafka-go"
)

func InitListener(appCtx context.Context, configuration boilerplate.KafkaListenerConfiguration,
	notificationSender sender.ISender, userGoWrapper user_go.IUserGoWrapper) kafka_listener.IKafkaListener {
	return kafka_listener.NewSingleListener(configuration, kafka_listener.NewCommand("tokenomics notifications",
		func(executionData kafka_listener.ExecutionData, request ...kafka.Message) []kafka.Message {
			singleMessage := request[0]

			mapped, err := mapKafkaMessages(singleMessage)

			if err != nil {
				apm_helper.LogError(err, executionData.Context)

				return []kafka.Message{singleMessage}
			}

			result, err := process(*mapped, executionData.Context, notificationSender, userGoWrapper)

			if err != nil {
				apm_helper.LogError(err, executionData.Context)
			}

			successfulMessages := make([]kafka.Message, 0)

			if result != nil {
				successfulMessages = append(successfulMessages, *result)
			}

			return successfulMessages
		}, true), appCtx)
}
