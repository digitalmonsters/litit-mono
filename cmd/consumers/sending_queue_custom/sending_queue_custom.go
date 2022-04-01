package sending_queue_custom

import (
	"context"
	"github.com/digitalmonsters/go-common/wrappers/notification_handler"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	notificationPkg "github.com/digitalmonsters/notification-handler/pkg/notification"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/segmentio/kafka-go"
	"time"
)

func process(event newCustomSendingEvent, ctx context.Context, notifySender sender.ISender) (*kafka.Message, error) {
	db := database.GetDbWithContext(database.DbTypeMaster, ctx)

	tx := db.Begin()
	defer tx.Rollback()

	if err := tx.Create(&database.Notification{
		UserId:    event.UserId,
		Type:      database.GetNotificationType("custom_reward_increase"),
		Title:     event.Title,
		Message:   event.Body,
		CreatedAt: time.Now().UTC(),
	}).Error; err != nil {
		return nil, err
	}
	if err := notificationPkg.IncrementUnreadNotificationsCounter(tx, event.UserId); err != nil {
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	_, err := notifySender.SendCustomTemplateToUser(notification_handler.NotificationChannelPush,
		event.UserId, "custom", "popup", event.Title, event.Body, event.Headline, nil, ctx)

	if err != nil {
		return nil, err
	}
	return &event.Messages, nil
}
