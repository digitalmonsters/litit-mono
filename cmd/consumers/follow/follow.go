package follow

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
)

func process(event newSendingEvent, ctx context.Context, notifySender sender.ISender,
	userGoWrapper user_go.IUserGoWrapper) (*kafka.Message, error) {
	if !event.Follow {
		return &event.Messages, nil
	}

	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "user_id", event.UserId)
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "to_user_id", event.ToUserId)

	var userData user_go.UserRecord
	var err error

	resp := <-userGoWrapper.GetUsers([]int64{event.UserId}, ctx, false)
	if resp.Error != nil {
		return nil, resp.Error.ToError()
	}

	var ok bool
	if userData, ok = resp.Response[event.UserId]; !ok {
		return &event.Messages, errors.WithStack(errors.New("user not found")) // we should continue, no need to retry
	}

	firstName, lastName := userData.GetFirstAndLastNameWithPrivacy()

	renderingVariables := database.RenderingVariables{
		"firstname": firstName,
		"lastname":  lastName,
	}

	shouldRetry, err := notifySender.PushNotification(database.Notification{
		UserId:             event.ToUserId,
		Type:               "push.profile.following",
		RelatedUserId:      null.IntFrom(event.UserId),
		RenderingVariables: renderingVariables,
		CustomData:         database.CustomData{"user_id": event.UserId},
	}, event.UserId, 0, "follow", userData.Language, "user_follow", ctx)
	if err != nil {
		if shouldRetry {
			return nil, errors.WithStack(err)
		}

		return &event.Messages, err
	}

	return &event.Messages, nil
}
