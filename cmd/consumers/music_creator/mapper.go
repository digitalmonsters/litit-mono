package music_creator

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
)

type newSendingEvent struct {
	eventsourcing.MusicCreatorModel
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
