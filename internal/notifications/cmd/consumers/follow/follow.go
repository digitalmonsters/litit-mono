package follow

import (
	"context"

	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/digitalmonsters/notification-handler/pkg/utils"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
)

func process(event newSendingEvent, ctx context.Context, notifySender sender.ISender) (*kafka.Message, error) {
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "user_id", event.UserId)
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "to_user_id", event.ToUserId)
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "follow", event.Follow)

	templateName := "follow"

	if !event.Follow {
		if err := notifySender.UnapplyEvent(event.ToUserId, templateName, event.UserId, 0, ctx); err != nil {
			return nil, errors.WithStack(err)
		}

		return &event.Messages, nil
	}

	renderData, language, err := utils.GetUserRenderingVariablesWithLanguage(event.UserId, ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	shouldRetry, err := notifySender.PushNotification(database.Notification{
		UserId:             event.ToUserId,
		Type:               "push.profile.following",
		RelatedUserId:      null.IntFrom(event.UserId),
		RenderingVariables: renderData,
		CustomData:         database.CustomData{"user_id": event.UserId},
	}, "", event.UserId, 0, templateName, language, "user_follow", ctx)
	if err != nil {
		if shouldRetry {
			return nil, errors.WithStack(err)
		}

		return &event.Messages, err
	}

	return &event.Messages, nil
}
