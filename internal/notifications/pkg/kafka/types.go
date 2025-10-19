package kafka

import (
	"github.com/digitalmonsters/go-common/eventsourcing"
)

type KafkaMessage struct {
	EventType eventsourcing.EmailNotificationType `json:"event_type"`
	Payload   map[string]interface{}              `json:"payload"`
}
