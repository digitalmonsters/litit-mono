package dislike

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/segmentio/kafka-go"
)

type dislikeCount struct {
	Count    int64
	Messages []kafka.Message `json:"-"`
}

func mapKafkaMessages(messages []kafka.Message) (map[int64]*dislikeCount, []error, []kafka.Message) {
	eventsMap := make(map[int64]*dislikeCount)
	var appErrors []error
	var msg []kafka.Message

	for _, message := range messages {
		var event eventsourcing.ContentDislikeEventData

		if err := json.Unmarshal(message.Value, &event); err != nil {
			appErrors = append(appErrors, err)
			msg = append(msg, message)
			continue
		}

		if _, ok := eventsMap[event.Id]; !ok {
			eventsMap[event.Id] = &dislikeCount{}
		}

		val := eventsMap[event.Id]

		val.Messages = append(val.Messages, message)
		val.Count = event.Dislikes
	}

	return eventsMap, appErrors, msg
}
