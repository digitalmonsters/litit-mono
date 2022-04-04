package comment

import (
	"context"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/kafka_listener"
	"github.com/digitalmonsters/go-common/wrappers/comment"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/segmentio/kafka-go"
)

func InitListener(appCtx context.Context, configuration boilerplate.KafkaListenerConfiguration,
	notificationSender sender.ISender, userGoWrapper user_go.IUserGoWrapper, contentWrapper content.IContentWrapper,
	commentWrapper comment.ICommentWrapper) kafka_listener.IKafkaListener {
	return kafka_listener.NewSingleListener(configuration, kafka_listener.NewCommand("comments",
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

			result, err := process(*mapped, executionData.Context, notificationSender, userGoWrapper, contentWrapper,
				commentWrapper, executionData.ApmTransaction)

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
