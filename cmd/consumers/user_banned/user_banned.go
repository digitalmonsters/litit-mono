package user_banned

import (
	"context"
	"time"

	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
)

func process(event newSendingEvent, ctx context.Context, notifySender sender.ISender) (*kafka.Message, error) {
	if event.CrudOperation != eventsourcing.ChangeEventTypeUpdated && event.CrudOperationReason != "ban" {
		return &event.Messages, nil
	}

	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "crud_operation_reason", event.BaseChangeEvent.CrudOperationReason)
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "crud_operation", event.BaseChangeEvent.CrudOperation)
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "user_id", event.UserId)

	if !event.BannedTill.Valid || !event.BannedTill.Time.After(time.Now().UTC()) {
		return &event.Messages, nil
	}

	shouldRetry, err := notifySender.PushNotification(database.Notification{
		UserId:        event.UserId,
		Type:          "popup",
		RelatedUserId: null.IntFrom(event.UserId),
	}, "", 0, 0, "user_banned", event.Language, "popup", ctx)
	if err != nil {
		if shouldRetry {
			return nil, errors.WithStack(err)
		}

		return &event.Messages, err
	}

	return &event.Messages, nil
}
