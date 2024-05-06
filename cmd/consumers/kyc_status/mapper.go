package kyc_status

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"gopkg.in/guregu/null.v4"
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
	Type   LegacyEventType `json:"type"`
	Reason null.String     `json:"reason"`
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
				event.CrudOperationReason = "kyc_status_updated"
			}
			if len(event.CrudOperationReason) <= 0 {
				var reason = lEvent.Reason.ValueOrZero()
				if len(event.KycReason) > 0 {
					event.KycReason = eventsourcing.KycReason(reason)
				}
			}
		}
	}

	event.Messages = message

	return &event, nil
}
