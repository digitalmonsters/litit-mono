package sending_queue

import (
	"context"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
	"strconv"
)

func process(event newSendingEvent, ctx context.Context, notifySender sender.ISender, userGoWrapper user_go.IUserGoWrapper) (*kafka.Message, error) {
	db := database.GetDbWithContext(database.DbTypeMaster, ctx)
	tx := db.Begin()
	defer tx.Rollback()

	resp := <-userGoWrapper.GetUsers([]int64{event.UserId}, ctx, false)
	if resp.Error != nil {
		return nil, resp.Error.ToError()
	}

	var userData user_go.UserRecord
	var ok bool

	if userData, ok = resp.Response[event.UserId]; !ok {
		return &event.Messages, errors.WithStack(errors.New("user not found")) // we should continue, no need to retry
	}

	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "user_id", event.UserId)
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "template_name", event.TemplateName)

	var err error

	if err != nil {
		return nil, errors.WithStack(err)
	}
	var relatedUserId null.Int
	var isCustomTemplate bool

	switch event.TemplateName {
	case "other_referrals_joined":
		fallthrough
	case "first_referral_joined":
		isCustomTemplate = true
		relatedUserId = GetRelatedUserIdFromRenderData(event.RenderingVariables, "referral_id", ctx)
	case "referral_greeting":
		isCustomTemplate = true
		relatedUserId = GetRelatedUserIdFromRenderData(event.RenderingVariables, "referrer_id", ctx)
	}

	customData := event.CustomData

	if isCustomTemplate && relatedUserId.Valid {
		customData = map[string]interface{}{}
		customData["user_id"] = relatedUserId.Int64
	}

	shouldRetry, err := notifySender.PushNotification(database.Notification{
		UserId:             event.UserId,
		Type:               database.GetNotificationType(event.TemplateName),
		RelatedUserId:      relatedUserId,
		RenderingVariables: event.RenderingVariables,
		CustomData:         customData,
	}, event.UserId, 0, event.TemplateName, userData.Language, "", ctx)
	if err != nil {
		if shouldRetry {
			return nil, errors.WithStack(err)
		}

		return &event.Messages, err
	}

	return &event.Messages, nil
}

func GetRelatedUserIdFromRenderData(renderingVariables map[string]string, relatedUser string, ctx context.Context) null.Int {
	var relatedUserId null.Int
	if val, ok := renderingVariables[relatedUser]; ok {
		parsed, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			apm_helper.LogError(err, ctx)
		} else {
			relatedUserId = null.IntFrom(parsed)
		}
	}
	return relatedUserId
}
