package tokenomics_notification

import (
	"context"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
)

func process(event newSendingEvent, ctx context.Context, notifySender sender.ISender, userGoWrapper user_go.IUserGoWrapper) (*kafka.Message, error) {
	var err error
	rendererData := map[string]string{}
	var templateName string
	var templateType string
	contentId := null.Int{}

	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "event_type", string(event.Type))
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "user_id", event.Payload.UserId)
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "related_user_id", event.Payload.RelatedUserId.ValueOrZero())

	if event.Payload.PointsAmount.Valid {
		rendererData["pointsAmount"] = event.Payload.PointsAmount.Decimal.String()
	}

	var userData user_go.UserRecord
	var ok bool

	switch event.Type {
	case eventsourcing.TokenomicsNotificationTip:
		resp := <-userGoWrapper.GetUsers([]int64{event.Payload.RelatedUserId.ValueOrZero()}, ctx, false)
		if resp.Error != nil {
			return nil, resp.Error.ToError()
		}

		if userData, ok = resp.Response[event.Payload.RelatedUserId.ValueOrZero()]; !ok {
			return &event.Messages, errors.WithStack(errors.New("user not found")) // we should continue, no need to retry
		}

		templateName = "tip"
		templateType = "push.tip"

		firstName, lastName := userData.GetFirstAndLastNameWithPrivacy()

		rendererData["firstname"] = firstName
		rendererData["lastname"] = lastName

		contentId = null.IntFrom(0)
	case eventsourcing.TokenomicsNotificationDailyBonusTime:
		resp := <-userGoWrapper.GetUsers([]int64{event.Payload.UserId}, ctx, false)
		if resp.Error != nil {
			return nil, resp.Error.ToError()
		}

		if userData, ok = resp.Response[event.Payload.UserId]; !ok {
			return &event.Messages, errors.WithStack(errors.New("user not found")) // we should continue, no need to retry
		}

		templateName = "bonus_time"
		templateType = "push.bonus.daily"
	case eventsourcing.TokenomicsNotificationDailyBonusFollowers:
		resp := <-userGoWrapper.GetUsers([]int64{event.Payload.UserId}, ctx, false)
		if resp.Error != nil {
			return nil, resp.Error.ToError()
		}

		if userData, ok = resp.Response[event.Payload.UserId]; !ok {
			return &event.Messages, errors.WithStack(errors.New("user not found")) // we should continue, no need to retry
		}

		templateName = "bonus_followers"
		templateType = "push.bonus.followers"
	default:
		return &event.Messages, nil
	}

	shouldRetry, err := notifySender.PushNotification(database.Notification{
		UserId:             event.Payload.UserId,
		Type:               templateType,
		ContentId:          contentId,
		RelatedUserId:      event.Payload.RelatedUserId,
		RenderingVariables: rendererData,
	}, event.Payload.RelatedUserId.ValueOrZero(), 0, templateName, userData.Language, "default", ctx)
	if err != nil {
		if shouldRetry {
			return nil, errors.WithStack(err)
		}

		return &event.Messages, err
	}

	return &event.Messages, nil
}
