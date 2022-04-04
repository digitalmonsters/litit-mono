package kyc_status

import (
	"context"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/wrappers/notification_handler"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/notification"
	"github.com/digitalmonsters/notification-handler/pkg/renderer"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/segmentio/kafka-go"
	"time"
)

func process(event newSendingEvent, ctx context.Context, notifySender sender.ISender) (*kafka.Message, error) {
	if event.CrudOperation != eventsourcing.ChangeEventTypeUpdated || !(event.KycStatus == eventsourcing.KycStatusRejected || event.KycStatus == eventsourcing.KycStatusVerified) {
		return &event.Messages, nil
	}

	var err error
	var title string
	var body string
	var headline string
	var templateName string
	renderData := map[string]string{}

	db := database.GetDb(database.DbTypeMaster).WithContext(ctx)

	if event.KycStatus == eventsourcing.KycStatusVerified {
		templateName = "kyc_status_verified"
	} else if event.KycStatus == eventsourcing.KycStatusRejected {
		templateName = "kyc_status_rejected"
		renderData = map[string]string{
			"reason": event.CrudOperationReason,
		}
	} else {
		return &event.Messages, nil
	}

	title, body, headline, _, err = notifySender.RenderTemplate(db, templateName, renderData)
	if err == renderer.TemplateRenderingError {
		return &event.Messages, err // we should continue, no need to retry
	} else if err != nil {
		return nil, err
	}

	if _, err = notifySender.SendCustomTemplateToUser(notification_handler.NotificationChannelPush, event.UserId, templateName, "default",
		title, body, headline, nil, ctx); err != nil {
		return nil, err
	}

	reason := eventsourcing.KycReason(event.CrudOperationReason)
	if err = db.Create(&database.Notification{
		UserId:    event.UserId,
		Type:      "push.kyc.status",
		Title:     title,
		Message:   body,
		CreatedAt: time.Now().UTC(),
		KycReason: &reason,
		KycStatus: &event.KycStatus,
	}).Error; err != nil {
		return nil, err
	}

	if err = notification.IncrementUnreadNotificationsCounter(db, event.UserId); err != nil {
		return nil, err
	}

	return &event.Messages, nil
}
