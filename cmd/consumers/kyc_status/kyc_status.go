package kyc_status

import (
	"context"
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
	apm_helper.AddApmLabel(apmTransaction, "crud_operation_reason", event.BaseChangeEvent.CrudOperationReason)
	apm_helper.AddApmLabel(apmTransaction, "crud_operation", event.BaseChangeEvent.CrudOperation)
	apm_helper.AddApmLabel(apmTransaction, "user_id", event.UserId)

	if event.CrudOperationReason != "kyc_status_updated" {
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
			"reason": string(event.KycReason),
		}
	} else {
		return &event.Messages, nil
	}

	var template database.RenderTemplate
	title, body, headline, template, err = notifySender.RenderTemplate(db, templateName, renderData, event.Language)
	if err == renderer.TemplateRenderingError {
		return &event.Messages, err // we should continue, no need to retry
	} else if err != nil {
		return nil, err
	}

	customData := database.CustomData{"image_url": template.ImageUrl, "route": template.Route}
	if _, err = notifySender.SendCustomTemplateToUser(notification_handler.NotificationChannelPush, event.UserId, templateName, "default",
		title, body, headline, customData, ctx); err != nil {
		return nil, err
	}

	reason := event.KycReason

	var dbReason *eventsourcing.KycReason

	if event.KycStatus != eventsourcing.KycStatusVerified {
		dbReason = &reason
	}

	if err = db.Create(&database.Notification{
		UserId:             event.UserId,
		Type:               "push.kyc.status",
		Title:              title,
		Message:            body,
		CreatedAt:          time.Now().UTC(),
		KycStatus:          &event.KycStatus,
		RenderingVariables: renderData,
		KycReason:          dbReason,
		CustomData:         customData,
	}).Error; err != nil {
		return nil, err
	}

	if err = notification.IncrementUnreadNotificationsCounter(db, event.UserId); err != nil {
		return nil, err
	}

	return &event.Messages, nil
}
