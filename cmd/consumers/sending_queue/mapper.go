package sending_queue

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/wrappers/notification_handler"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
)

type newSendingEvent struct {
	notification_handler.SendNotificationWithTemplate
	Messages kafka.Message `json:"-"`
}

func mapKafkaMessages(message kafka.Message) (*newSendingEvent, error) {
	var event newSendingEvent

	if err := json.Unmarshal(message.Value, &event); err != nil {
		return nil, errors.WithStack(err)
	}

	event.Messages = message

	return &event, nil
}
