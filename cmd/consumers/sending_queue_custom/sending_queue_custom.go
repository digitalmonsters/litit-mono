package sending_queue_custom

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

func process(event newCustomSendingEvent, ctx context.Context, notifySender sender.ISender, apmTransaction *apm.Transaction) (*kafka.Message, error) {
	db := database.GetDbWithContext(database.DbTypeMaster, ctx)

	tx := db.Begin()
	defer tx.Rollback()

	apm_helper.AddApmLabel(apmTransaction, "user_id", event.UserId)

	nt := &database.Notification{
		UserId:    event.UserId,
		Type:      database.GetNotificationType("custom_reward_increase"),
		Title:     event.Title,
		Message:   event.Body,
		CreatedAt: time.Now().UTC(),
	}

	if err := tx.Create(nt).Error; err != nil {
		return nil, err
	}

	apm_helper.AddApmLabel(apmTransaction, "notification_id", nt.Id.String())

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
