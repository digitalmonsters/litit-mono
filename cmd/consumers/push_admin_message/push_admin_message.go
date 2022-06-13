package push_admin_message

import (
	"context"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/wrappers/notification_handler"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	notificationPkg "github.com/digitalmonsters/notification-handler/pkg/notification"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
	"time"
)

func process(event newSendingEvent, ctx context.Context, notifySender sender.ISender, apmTransaction *apm.Transaction) (*kafka.Message, error) {
	var err error

	apm_helper.AddApmLabel(apmTransaction, "user_id", event.UserId)
	apm_helper.AddApmLabel(apmTransaction, "title", event.Title)

	db := database.GetDb(database.DbTypeMaster).WithContext(ctx)

	_, err = notifySender.SendCustomTemplateToUser(notification_handler.NotificationChannelPush, event.UserId, "admin_bulk", "default",
		event.Title, event.Message, "", event.CustomData, ctx)
	if err != nil {
		return nil, err
	}

	nt := &database.Notification{
		UserId:     event.UserId,
		Type:       "push.admin.bulk",
		Title:      event.Title,
		Message:    event.Message,
		CreatedAt:  time.Now().UTC(),
		CustomData: event.CustomData,
	}

	if err = db.Create(nt).Error; err != nil {
		return nil, err
	}

	apm_helper.AddApmLabel(apmTransaction, "notification_id", nt.Id.String())

	if err = notificationPkg.IncrementUnreadNotificationsCounter(db, event.UserId); err != nil {
		return nil, err
	}

	return &event.Messages, nil
}
