package user_consumer

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
)

type wrappedEvent struct {
	eventsourcing.UserEvent
	Message kafka.Message `json:"-"`
}

func mapKafkaMessages(message kafka.Message) (*wrappedEvent, error) {
	var event wrappedEvent

	if err := json.Unmarshal(message.Value, &event); err != nil {
		return nil, errors.WithStack(err)
	}

	event.Message = message

	return &event, nil
}
