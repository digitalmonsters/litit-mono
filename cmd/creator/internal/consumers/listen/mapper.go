package listen

import (
	"encoding/json"
	"github.com/segmentio/kafka-go"
)

type listenCount struct {
	legacyEvent
	Messages []kafka.Message `json:"-"`
}

type legacyEvent struct {
	Id                int64 `json:"id"`
	Count             int64 `json:"count"`
	ListensCount      int64 `json:"listens_count"`
	ShortListensCount int64 `json:"short_listens_count"`
}

func mapKafkaMessages(messages []kafka.Message) (map[int64]*listenCount, []error, []kafka.Message) {
	eventsMap := make(map[int64]*listenCount)
	var appErrors []error
	var msg []kafka.Message

	for _, message := range messages {
		var event legacyEvent

		if err := json.Unmarshal(message.Value, &event); err != nil {
			appErrors = append(appErrors, err)
			msg = append(msg, message)
			continue
		}

		if event.ListensCount == 0 && event.ShortListensCount == 0 {
			continue
		}

		if _, ok := eventsMap[event.Id]; !ok {
			eventsMap[event.Id] = &listenCount{}
		}

		val := eventsMap[event.Id]

		val.Messages = append(val.Messages, message)
		val.Count = event.Count
		val.ListensCount = event.ListensCount
		val.ShortListensCount = event.ShortListensCount
	}

	return eventsMap, appErrors, msg
}
