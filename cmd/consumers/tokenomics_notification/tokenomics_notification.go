package tokenomics_notification

import (
	"context"

	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/translation"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/digitalmonsters/notification-handler/pkg/utils"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
)

func process(event newSendingEvent, ctx context.Context, notifySender sender.ISender) (*kafka.Message, error) {
	var err error
	var renderData database.RenderingVariables
	var templateName string
	var templateType string
	contentId := null.Int{}

	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "event_type", string(event.Type))
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "user_id", event.Payload.UserId)
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "related_user_id", event.Payload.RelatedUserId.ValueOrZero())

	var language translation.Language

	switch event.Type {
	case eventsourcing.TokenomicsNotificationTip:
		renderData, language, err = utils.GetUserRenderingVariablesWithLanguage(event.Payload.RelatedUserId.ValueOrZero(), ctx)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		templateName = "tip"
		templateType = "push.tip"

		contentId = null.IntFrom(0)
	case eventsourcing.TokenomicsNotificationDailyBonusTime:
		renderData, language, err = utils.GetUserRenderingVariablesWithLanguage(event.Payload.UserId, ctx)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		templateName = "bonus_time"
		templateType = "push.bonus.daily"
	case eventsourcing.TokenomicsNotificationDailyBonusFollowers:
		renderData, language, err = utils.GetUserRenderingVariablesWithLanguage(event.Payload.UserId, ctx)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		templateName = "bonus_followers"
		templateType = "push.bonus.followers"
	default:
		return &event.Messages, nil
	}

	if event.Payload.PointsAmount.Valid {
		renderData["pointsAmount"] = event.Payload.PointsAmount.Decimal.String()
	}

	shouldRetry, err := notifySender.PushNotification(database.Notification{
		UserId:             event.Payload.UserId,
		Type:               templateType,
		ContentId:          contentId,
		RelatedUserId:      event.Payload.RelatedUserId,
		RenderingVariables: renderData,
	}, "", event.Payload.RelatedUserId.ValueOrZero(), 0, templateName, language, "default", ctx)
	if err != nil {
		if shouldRetry {
			return nil, errors.WithStack(err)
		}

		return &event.Messages, err
	}

	return &event.Messages, nil
}
