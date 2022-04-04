package kyc_status

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
)

type newSendingEvent struct {
	eventsourcing.UserEvent
	Messages kafka.Message `json:"-"`
}

type LegacyEventType string

const (
	LegacyEventTypeKycStatusUpdated LegacyEventType = "kyc_status_updated"
)

type legacyEvent struct {
	Type LegacyEventType `json:"type"`
}

func mapKafkaMessages(message kafka.Message) (*newSendingEvent, error) {
	var event newSendingEvent
	var lEvent legacyEvent

	if err := json.Unmarshal(message.Value, &event); err != nil {
		return nil, errors.WithStack(err)
	}

	if err := json.Unmarshal(message.Value, &lEvent); err != nil {
		return nil, errors.WithStack(err)
	}

	if len(lEvent.Type) > 0 {
		switch lEvent.Type {
		case LegacyEventTypeKycStatusUpdated:
			if len(event.CrudOperation) <= 0 {
				event.CrudOperation = eventsourcing.ChangeEventTypeUpdated
			}
			if len(event.CrudOperationReason) <= 0 {
				event.CrudOperationReason = string(LegacyEventTypeKycStatusUpdated)
			}
		}
	}

	event.Messages = message

	return &event, nil
}
