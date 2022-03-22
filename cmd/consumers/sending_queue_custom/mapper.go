package sending_queue_custom

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/wrappers/notification_handler"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
)

type newCustomSendingEvent struct {
	notification_handler.SendNotificationWithCustomTemplate
	Messages kafka.Message `json:"-"`
}

func mapKafkaMessages(message kafka.Message) (*newCustomSendingEvent, error) {
	var event newCustomSendingEvent

	if err := json.Unmarshal(message.Value, &event); err != nil {
		return nil, errors.WithStack(err)
	}

	event.Messages = message

	return &event, nil
}
