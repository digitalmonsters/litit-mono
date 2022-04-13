package sending_queue

import (
	"context"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/wrappers/notification_handler"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	notificationPkg "github.com/digitalmonsters/notification-handler/pkg/notification"
	"github.com/digitalmonsters/notification-handler/pkg/renderer"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
	"strconv"
	"time"
)

func process(event newSendingEvent, ctx context.Context, notifySender sender.ISender, apmTransaction *apm.Transaction) (*kafka.Message, error) {
	db := database.GetDbWithContext(database.DbTypeMaster, ctx)
	tx := db.Begin()
	defer tx.Rollback()
	title, body, headline, renderingTemplate, err := notifySender.RenderTemplate(tx,
		event.TemplateName, event.RenderingVariables)

	apm_helper.AddApmLabel(apmTransaction, "user_id", event.UserId)
	apm_helper.AddApmLabel(apmTransaction, "template_name", event.TemplateName)

	if err != nil {
		return nil, errors.WithStack(err)
	}
	var relatedUserId null.Int
	var isReferralsTemplate = event.TemplateName == "other_referrals_joined" || event.TemplateName == "first_referral_joined"

	if isReferralsTemplate {
		if val, ok := event.RenderingVariables["referral_id"]; ok {
			parsed, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				apm_helper.CaptureApmError(err, apmTransaction)
			} else {
				relatedUserId = null.IntFrom(parsed)
			}
		}
	}
	nf := &database.Notification{
		UserId:             event.UserId,
		Type:               database.GetNotificationType(event.TemplateName),
		Title:              title,
		Message:            body,
		CreatedAt:          time.Now().UTC(),
		RenderingVariables: event.RenderingVariables,
		RelatedUserId:      relatedUserId,
	}

	if err := tx.Create(nf).Error; err != nil {
		return nil, err
	}

	apm_helper.AddApmLabel(apmTransaction, "notification_id", nf.Id.String())

	if err := notificationPkg.IncrementUnreadNotificationsCounter(tx, event.UserId); err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	if isReferralsTemplate {
		var customData = map[string]interface{}{}

		if relatedUserId.Valid {
			customData["user_id"] = relatedUserId.Int64
		}
		_, err = notifySender.SendCustomTemplateToUser(notification_handler.NotificationChannelPush, event.UserId,
			renderingTemplate.Id, renderingTemplate.Kind, title, body, headline, customData, ctx)
	} else {
		_, err = notifySender.SendTemplateToUser(notification_handler.NotificationChannelPush,
			title, body, headline, renderingTemplate, event.UserId, event.RenderingVariables, ctx)
	}
	if err == renderer.TemplateRenderingError {
		return &event.Messages, err // we should continue, no need to retry
	}
	if err != nil {
		return nil, err
	}
	return &event.Messages, nil
}
