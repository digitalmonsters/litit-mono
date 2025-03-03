package push_admin_message

import (
	"context"

	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/translation"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
)

func process(event newSendingEvent, ctx context.Context, notifySender sender.ISender) (*kafka.Message, error) {
	var err error

	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "user_id", event.UserId)
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "title", event.Title)

	if event.CustomData == nil {
		event.CustomData = database.CustomData{}
	}

	shouldRetry, err := notifySender.PushNotification(database.Notification{
		UserId:     event.UserId,
		Type:       "push.admin.bulk",
		Title:      event.Title,
		Message:    event.Message,
		CustomData: event.CustomData,
	}, "", event.UserId, 0, "push_admin", translation.DefaultUserLanguage, "default", ctx)
	if err != nil {
		if shouldRetry {
			return nil, errors.WithStack(err)
		}

		return &event.Messages, err
	}

	return &event.Messages, nil
}
