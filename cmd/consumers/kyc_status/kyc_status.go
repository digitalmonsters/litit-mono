package kyc_status

import (
	"context"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
)

func process(event newSendingEvent, ctx context.Context, notifySender sender.ISender) (*kafka.Message, error) {
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "crud_operation_reason", event.BaseChangeEvent.CrudOperationReason)
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "crud_operation", event.BaseChangeEvent.CrudOperation)
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "user_id", event.UserId)

	if event.CrudOperationReason != "kyc_status_updated" {
		return &event.Messages, nil
	}

	var err error
	var templateName string
	renderData := map[string]string{}

	if event.KycStatus == eventsourcing.KycStatusVerified {
		templateName = "kyc_status_verified"
	} else if event.KycStatus == eventsourcing.KycStatusRejected {
		templateName = "kyc_status_rejected"
		renderData = map[string]string{
			"reason": string(event.KycReason),
		}
	} else {
		return &event.Messages, nil
	}

	reason := event.KycReason

	var dbReason *eventsourcing.KycReason

	if event.KycStatus != eventsourcing.KycStatusVerified {
		dbReason = &reason
	}

	shouldRetry, err := notifySender.PushNotification(database.Notification{
		UserId:             event.UserId,
		Type:               "push.kyc.status",
		KycStatus:          &event.KycStatus,
		KycReason:          dbReason,
		RenderingVariables: renderData,
	}, event.UserId, 0, templateName, event.Language, "default", ctx)
	if err != nil {
		if shouldRetry {
			return nil, errors.WithStack(err)
		}

		return &event.Messages, err
	}

	return &event.Messages, nil
}
