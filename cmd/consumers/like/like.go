package like

import (
	"context"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/digitalmonsters/go-common/wrappers/notification_handler"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	notificationPkg "github.com/digitalmonsters/notification-handler/pkg/notification"
	"github.com/digitalmonsters/notification-handler/pkg/renderer"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
	"time"
)

func process(event newSendingEvent, ctx context.Context, notifySender sender.ISender, userGoWrapper user_go.IUserGoWrapper,
	contentWrapper content.IContentWrapper, apmTransaction *apm.Transaction) (*kafka.Message, error) {
	if !event.Like || event.ContentAuthorId == event.UserId {
		return &event.Messages, nil
	}

	var userData user_go.UserRecord
	var err error

	resp := <-userGoWrapper.GetUsers([]int64{event.UserId}, apmTransaction, false)
	if resp.Error != nil {
		return nil, resp.Error.ToError()
	}

	var ok bool
	if userData, ok = resp.Items[event.UserId]; !ok {
		return &event.Messages, errors.WithStack(errors.New("user not found")) // we should continue, no need to retry
	}

	var title string
	var body string
	var headline string

	db := database.GetDb(database.DbTypeMaster).WithContext(ctx)

	title, body, headline, _, err = notifySender.RenderTemplate(db, "content_like", map[string]string{
		"firstname": userData.Firstname,
		"lastname":  userData.Lastname,
	})
	if err == renderer.TemplateRenderingError {
		return &event.Messages, err // we should continue, no need to retry
	} else if err != nil {
		return nil, err
	}

	if _, err = notifySender.SendCustomTemplateToUser(notification_handler.NotificationChannelPush, event.ContentAuthorId,
		title, body, headline, ctx); err != nil {
		return nil, err
	}

	contentResp := <-contentWrapper.GetInternal([]int64{event.ContentId}, false, apmTransaction, false)
	if contentResp.Error != nil {
		return nil, err
	}

	notification := database.Notification{
		UserId:        event.ContentAuthorId,
		Type:          "push.content.like",
		Title:         title,
		Message:       body,
		CreatedAt:     time.Now().UTC(),
		ContentId:     null.IntFrom(event.ContentId),
		RelatedUserId: null.IntFrom(event.UserId),
	}

	var contentData content.SimpleContent

	if contentData, ok = contentResp.Items[event.ContentId]; ok {
		notification.Content = &database.NotificationContent{
			Id:      event.ContentId,
			Width:   contentData.Width,
			Height:  contentData.Height,
			VideoId: contentData.VideoId,
		}
	}

	if err = db.Create(&notification).Error; err != nil {
		return nil, err
	}

	if err = notificationPkg.IncrementUnreadNotificationsCounter(db, event.UserId); err != nil {
		return nil, err
	}

	return &event.Messages, nil
}
