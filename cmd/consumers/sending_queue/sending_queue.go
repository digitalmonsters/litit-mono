package sending_queue

import (
	"context"
	"github.com/digitalmonsters/go-common/wrappers/notification_handler"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	notificationPkg "github.com/digitalmonsters/notification-handler/pkg/notification"
	"github.com/digitalmonsters/notification-handler/pkg/renderer"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"time"
)

func process(event newSendingEvent, ctx context.Context, notifySender sender.ISender) (*kafka.Message, error) {
	db := database.GetDbWithContext(database.DbTypeMaster, ctx)
	tx := db.Begin()
	defer tx.Rollback()
	title, body, headline, renderingTemplate, err := notifySender.RenderTemplate(tx,
		event.TemplateName, event.RenderingVariables)

	if err != nil {
		return nil, errors.WithStack(err)
	}
	if err := tx.Create(&database.Notification{
		UserId:    event.UserId,
		Type:      database.GetNotificationType(event.TemplateName),
		Title:     title,
		Message:   body,
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
	_, err = notifySender.SendTemplateToUser(notification_handler.NotificationChannelPush,
		title, body, headline, renderingTemplate, event.UserId, event.RenderingVariables, ctx)

	if err == renderer.TemplateRenderingError {
		return &event.Messages, err // we should continue, no need to retry
	}
	if err != nil {
		return nil, err
	}
	return &event.Messages, nil
}
