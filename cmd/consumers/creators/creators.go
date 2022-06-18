package creators

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
)

func process(event newSendingEvent, ctx context.Context, notifySender sender.ISender, userGoWrapper user_go.IUserGoWrapper) (*kafka.Message, error) {
	var err error

	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "user_id", event.UserId)
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "creator_id", event.Id)

	renderingData := map[string]string{
		"status": fmt.Sprint(event.Status),
	}
	var templateName string

	switch event.Status {
	case user_go.CreatorStatusRejected:
		templateName = "creator_status_rejected"
	case user_go.CreatorStatusApproved:
		templateName = "creator_status_approved"
	case user_go.CreatorStatusPending:
		templateName = "creator_status_pending"
	default:
		return &event.Messages, nil
	}

	resp := <-userGoWrapper.GetUsers([]int64{event.UserId}, ctx, false)
	if resp.Error != nil {
		return nil, resp.Error.ToError()
	}

	var userData user_go.UserRecord
	var ok bool

	if userData, ok = resp.Response[event.UserId]; !ok {
		return &event.Messages, errors.WithStack(errors.New("user not found")) // we should continue, no need to retry
	}

	shouldRetry, err := notifySender.PushNotification(database.Notification{
		UserId:               event.UserId,
		Type:                 "push.content-creator.status",
		ContentCreatorStatus: &event.Status,
		RenderingVariables:   renderingData,
	}, event.UserId, 0, templateName, userData.Language, "content_creator", ctx)
	if err != nil {
		if shouldRetry {
			return nil, errors.WithStack(err)
		}

		return &event.Messages, err
	}

	return &event.Messages, nil
}
