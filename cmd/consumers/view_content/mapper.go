package view_content

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
)

type fullEvent struct {
	eventsourcing.ViewEvent
	Messages kafka.Message
}

func mapKafkaMessages(message kafka.Message) (*fullEvent, error) {
	var event fullEvent

	if err := json.Unmarshal(message.Value, &event); err != nil {
		return nil, errors.WithStack(err)
	}

	event.Messages = message

	return &event, nil
}
