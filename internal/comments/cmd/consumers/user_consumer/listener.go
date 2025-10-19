package user_consumer

import (
	"context"
	"github.com/digitalmonsters/comments/cmd/api/comments/notifiers/comment"
	"github.com/digitalmonsters/comments/cmd/api/comments/notifiers/content_comments_counter"
	"github.com/digitalmonsters/comments/cmd/api/comments/notifiers/user_comments_counter"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/kafka_listener"
	"github.com/segmentio/kafka-go"
)

func InitListener(applicationCtx context.Context, kafkaConfig boilerplate.KafkaListenerConfiguration,
	commentNotifier *comment.Notifier, contentCommentsNotifier *content_comments_counter.Notifier,
	userCommentsNotifier *user_comments_counter.Notifier) kafka_listener.IKafkaListener {
	return kafka_listener.NewSingleListener(kafkaConfig, kafka_listener.NewCommand("user event listener",
		func(executionData kafka_listener.ExecutionData, request ...kafka.Message) []kafka.Message {
			singleMessage := request[0]

			mapped, err := mapKafkaMessages(singleMessage)

			if err != nil {
				apm_helper.LogError(err, executionData.Context)

				return []kafka.Message{singleMessage}
			}

			result, err := process(*mapped, executionData.ApmTransaction, executionData.Context,
				commentNotifier, contentCommentsNotifier, userCommentsNotifier)

			if err != nil {
				apm_helper.LogError(err, executionData.Context)
			}

			successfulMessages := make([]kafka.Message, 0)

			if result != nil {
				successfulMessages = append(successfulMessages, *result)
			}

			return successfulMessages
		}, true), applicationCtx)
}
