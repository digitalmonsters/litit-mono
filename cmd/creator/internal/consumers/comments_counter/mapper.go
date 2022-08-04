package comments_counter

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
)

type eventData struct {
	eventsourcing.CommentCountOnContentEvent
	Messages kafka.Message `json:"-"`
}

func mapKafkaMessages(message kafka.Message) (*eventData, error) {
	var event eventData

	if err := json.Unmarshal(message.Value, &event); err != nil {
		return nil, errors.WithStack(err)
	}

	event.Messages = message

	return &event, nil
}
