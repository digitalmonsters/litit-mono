package shares

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/segmentio/kafka-go"
)

type sharesCount struct {
	eventsourcing.ContentEvent
	Messages []kafka.Message `json:"-"`
}

func mapKafkaMessages(messages []kafka.Message) (map[int64]*sharesCount, []error, []kafka.Message) {
	eventsMap := make(map[int64]*sharesCount)
	var appErrors []error
	var msg []kafka.Message

	for _, message := range messages {
		var event sharesCount

		if err := json.Unmarshal(message.Value, &event); err != nil {
			appErrors = append(appErrors, err)
			msg = append(msg, message)
			continue
		}

		if shouldExecute(event.ContentEvent) {
			if _, ok := eventsMap[event.Id]; !ok {
				eventsMap[event.Id] = &sharesCount{}
			}

			val := eventsMap[event.Id]
			val.Messages = append(val.Messages, message)
			val.ContentEvent = event.ContentEvent
		} else {
			msg = append(msg, message) //ignore
		}
	}

	return eventsMap, appErrors, msg
}

func shouldExecute(event eventsourcing.ContentEvent) bool {
	return event.BaseChangeEvent.CrudOperation == eventsourcing.ChangeEventTypeUpdated &&
		event.BaseChangeEvent.CrudOperationReason == "shares_counter_updated" &&
		event.ContentType == eventsourcing.ContentTypeMusic
}
