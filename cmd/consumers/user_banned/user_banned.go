package user_banned

import (
	"context"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/wrappers/notification_handler"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	notificationPkg "github.com/digitalmonsters/notification-handler/pkg/notification"
	"github.com/digitalmonsters/notification-handler/pkg/renderer"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
	"time"
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

	db := database.GetDb(database.DbTypeMaster).WithContext(ctx)

	var err error

	var title string
	var body string
	var headline string

	var template = "user_banned"

	title, body, headline, _, err = notifySender.RenderTemplate(db, template, map[string]string{}, event.Language)
	if err == renderer.TemplateRenderingError {
		return &event.Messages, err // we should continue, no need to retry
	} else if err != nil {
		return nil, err
	}

	if _, err = notifySender.SendCustomTemplateToUser(notification_handler.NotificationChannelPush, event.UserId, template, "popup",
		title, body, headline, nil, ctx); err != nil {
		return nil, err
	}

	notification := database.Notification{
		UserId:    event.UserId,
		Type:      "popup",
		Title:     title,
		Message:   body,
		CreatedAt: time.Now().UTC(),
	}

	if err = db.Create(&notification).Error; err != nil {
		return nil, err
	}

	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "notification_id", notification.Id.String())

	if err = notificationPkg.IncrementUnreadNotificationsCounter(db, event.UserId); err != nil {
		return nil, err
	}

	return &event.Messages, nil
}
