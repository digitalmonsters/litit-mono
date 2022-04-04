package creators

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/wrappers/notification_handler"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/notification"
	"github.com/digitalmonsters/notification-handler/pkg/renderer"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
	"time"
)

func process(event newSendingEvent, ctx context.Context, notifySender sender.ISender, apmTransaction *apm.Transaction) (*kafka.Message, error) {
	var err error

	apm_helper.AddApmLabel(apmTransaction, "user_id", event.UserId)
	apm_helper.AddApmLabel(apmTransaction, "creator_id", event.Id)

	renderingData := map[string]string{
		"status": fmt.Sprint(event.Status),
	}

	db := database.GetDb(database.DbTypeMaster).WithContext(ctx)

	var title string
	var body string
	var headline string
	var templateName string

	switch event.Status {
	case eventsourcing.CreatorStatusRejected:
		templateName = "creator_status_rejected"
	case eventsourcing.CreatorStatusApproved:
		templateName = "creator_status_approved"
	case eventsourcing.CreatorStatusPending:
		templateName = "creator_status_pending"
	default:
		return &event.Messages, nil
	}

	title, body, headline, _, err = notifySender.RenderTemplate(db, templateName, renderingData)
	if err == renderer.TemplateRenderingError {
		return &event.Messages, err // we should continue, no need to retry
	} else if err != nil {
		return nil, err
	}

	if _, err = notifySender.SendCustomTemplateToUser(notification_handler.NotificationChannelPush, event.UserId, templateName, "content_creator",
		title, body, headline, nil, ctx); err != nil {
		return nil, err
	}

	if err = db.Create(&database.Notification{
		UserId:               event.UserId,
		Type:                 "push.content-creator.status",
		Title:                title,
		Message:              body,
		CreatedAt:            time.Now().UTC(),
		ContentCreatorStatus: &event.Status,
	}).Error; err != nil {
		return nil, err
	}

	if err = notification.IncrementUnreadNotificationsCounter(db, event.UserId); err != nil {
		return nil, err
	}

	return &event.Messages, nil
}
