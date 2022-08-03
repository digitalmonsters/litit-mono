package view_content

import (
	"context"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/kafka_listener"
	"github.com/digitalmonsters/go-common/wrappers/go_tokenomics"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

func InitListener(applicationCtx context.Context, db *gorm.DB, configuration boilerplate.KafkaListenerConfiguration,
	goTokenomicsWrapper go_tokenomics.IGoTokenomicsWrapper) kafka_listener.IKafkaListener {
	handler := NewViewContentService(goTokenomicsWrapper)

	l := kafka_listener.NewSingleListener(configuration, kafka_listener.NewCommand("view content",
		func(executionData kafka_listener.ExecutionData, request ...kafka.Message) []kafka.Message {
			singleMessage := request[0]

			mapped, err := mapKafkaMessages(singleMessage)

			if err != nil {
				apm_helper.LogError(err, executionData.Context)

				return []kafka.Message{singleMessage}
			}

			result := handler.process(db.WithContext(executionData.Context), *mapped, executionData.Context)

			successfulMessages := make([]kafka.Message, 0)

			if result != nil {
				successfulMessages = append(successfulMessages, *result)
			}

			return successfulMessages
		}, true), applicationCtx)

	return l
}
