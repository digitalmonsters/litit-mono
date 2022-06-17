package like

import (
	"context"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
)

func process(event newSendingEvent, ctx context.Context, notifySender sender.ISender, userGoWrapper user_go.IUserGoWrapper,
	contentWrapper content.IContentWrapper) (*kafka.Message, error) {

	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "user_id", event.UserId)
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "content_id", event.ContentId)

	if !event.Like || event.ContentAuthorId == event.UserId {
		return &event.Messages, nil
	}

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

	notification := database.Notification{
		UserId:             event.ContentAuthorId,
		Type:               "push.content.like",
		ContentId:          null.IntFrom(event.ContentId),
		RelatedUserId:      null.IntFrom(event.UserId),
		RenderingVariables: renderingVariables,
	}

	var contentData content.SimpleContent

	contentResp := <-contentWrapper.GetInternal([]int64{event.ContentId}, false, apm.TransactionFromContext(ctx), false)
	if contentResp.Error != nil {
		return nil, err
	}

	if contentData, ok = contentResp.Response[event.ContentId]; ok {
		notification.Content = &database.NotificationContent{
			Id:      event.ContentId,
			Width:   contentData.Width,
			Height:  contentData.Height,
			VideoId: contentData.VideoId,
		}
	}

	shouldRetry, err := notifySender.PushNotification(notification, event.ContentId, event.UserId, "content_like", userData.Language, "default", ctx)
	if err != nil {
		if shouldRetry {
			return nil, errors.WithStack(err)
		}

		return &event.Messages, err
	}

	return &event.Messages, nil
}
