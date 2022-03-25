package push_admin_message

import (
	"context"
	"github.com/digitalmonsters/go-common/wrappers/notification_handler"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	notificationPkg "github.com/digitalmonsters/notification-handler/pkg/notification"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/segmentio/kafka-go"
	"time"
)

func process(event newSendingEvent, ctx context.Context, notifySender sender.ISender) (*kafka.Message, error) {
	var err error

	db := database.GetDb(database.DbTypeMaster).WithContext(ctx)

	_, err = notifySender.SendCustomTemplateToUser(notification_handler.NotificationChannelPush, event.UserId, event.Title, event.Message, "", ctx)
	if err != nil {
		return nil, err
	}

	if err = db.Create(&database.Notification{
		UserId:    event.UserId,
		Type:      "push.admin.bulk",
		Title:     event.Title,
		Message:   event.Message,
		CreatedAt: time.Now().UTC(),
	}).Error; err != nil {
		return nil, err
	}

	if err = notificationPkg.IncrementUnreadNotificationsCounter(db, event.UserId); err != nil {
		return nil, err
	}

	return &event.Messages, nil
}
