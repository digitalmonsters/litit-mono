package listened_music

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/segmentio/kafka-go"
)

type listenEvent struct {
	eventsourcing.ViewEvent
	Messages []kafka.Message `json:"-"`
}

func mapKafkaMessages(messages []kafka.Message) ([]*listenEvent, []kafka.Message, []error) {
	eventsFinal := make([]*listenEvent, 0)
	messagesToSkip := make([]kafka.Message, 0)
	var appErrors []error

	for _, message := range messages {
		var event listenEvent
		if err := json.Unmarshal(message.Value, &event); err != nil {
			appErrors = append(appErrors, err)
			messagesToSkip = append(messagesToSkip, message)
			continue
		}

		if event.ContentType != eventsourcing.ContentTypeMusic {
			messagesToSkip = append(messagesToSkip, message)
			continue
		}

		for _, e := range eventsFinal {
			if e.UserId == event.UserId && e.ContentId == event.ContentId {
				e.Messages = append(e.Messages, message)
			}
		}

		event.Messages = append(event.Messages, message)
		eventsFinal = append(eventsFinal, &event)
	}

	return eventsFinal, messagesToSkip, appErrors
}
