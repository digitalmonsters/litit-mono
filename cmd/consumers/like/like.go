package like

import (
	"context"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/digitalmonsters/notification-handler/pkg/utils"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
)

func process(event newSendingEvent, ctx context.Context, notifySender sender.ISender,
	contentWrapper content.IContentWrapper) (*kafka.Message, error) {
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "user_id", event.UserId)
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "content_id", event.ContentId)
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "like", event.Like)

	templateName := "content_like"

	if !event.Like || event.ContentAuthorId == event.UserId {
		if !event.Like {
			if err := notifySender.UnapplyEvent(event.ContentAuthorId, templateName, event.ContentId, 0, ctx); err != nil {
				return nil, errors.WithStack(err)
			}
		}

		return &event.Messages, nil
	}

	var err error

	renderData, language, err := utils.GetUserRenderingVariablesWithLanguage(event.UserId, ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	notification := database.Notification{
		UserId:             event.ContentAuthorId,
		Type:               "push.content.like",
		ContentId:          null.IntFrom(event.ContentId),
		RelatedUserId:      null.IntFrom(event.UserId),
		RenderingVariables: renderData,
	}

	var contentData content.SimpleContent
	var ok bool

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

	shouldRetry, err := notifySender.PushNotification(notification, event.ContentId, 0, "content_like", language, "default", ctx)
	if err != nil {
		if shouldRetry {
			return nil, errors.WithStack(err)
		}

		return &event.Messages, err
	}

	return &event.Messages, nil
}
