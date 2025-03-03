package creators

import (
	"context"
	"fmt"

	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/digitalmonsters/notification-handler/pkg/utils"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
)

func process(event newSendingEvent, ctx context.Context, notifySender sender.ISender) (*kafka.Message, error) {
	var err error

	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "user_id", event.UserId)
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "creator_id", event.Id)

	renderData, language, err := utils.GetUserRenderingVariablesWithLanguage(event.UserId, ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	renderData["status"] = fmt.Sprint(event.Status)
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

	shouldRetry, err := notifySender.PushNotification(database.Notification{
		UserId:               event.UserId,
		Type:                 "push.content-creator.status",
		ContentCreatorStatus: &event.Status,
		RenderingVariables:   renderData,
	}, "", event.UserId, 0, templateName, language, "content_creator", ctx)
	if err != nil {
		if shouldRetry {
			return nil, errors.WithStack(err)
		}

		return &event.Messages, err
	}

	return &event.Messages, nil
}
