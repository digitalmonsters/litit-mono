package follow

import (
	"context"
	"github.com/digitalmonsters/go-common/wrappers/notification_handler"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/notification"
	"github.com/digitalmonsters/notification-handler/pkg/renderer"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
	"time"
)

func process(event newSendingEvent, ctx context.Context, notifySender sender.ISender, userGoWrapper user_go.IUserGoWrapper,
	apmTransaction *apm.Transaction) (*kafka.Message, error) {
	if !event.Follow {
		return &event.Messages, nil
	}

	var userData user_go.UserRecord
	var err error
	var title string
	var body string
	var headline string

	resp := <-userGoWrapper.GetUsers([]int64{event.UserId}, apmTransaction, false)
	if resp.Error != nil {
		return nil, resp.Error.ToError()
	}

	var ok bool
	if userData, ok = resp.Items[event.UserId]; !ok {
		return &event.Messages, errors.WithStack(errors.New("user not found")) // we should continue, no need to retry
	}

	db := database.GetDb(database.DbTypeMaster).WithContext(ctx)

	title, body, headline, _, err = notifySender.RenderTemplate(db, "follow", map[string]string{
		"firstname": userData.Firstname,
		"lastname":  userData.Lastname,
	})
	if err == renderer.TemplateRenderingError {
		return &event.Messages, err // we should continue, no need to retry
	} else if err != nil {
		return nil, err
	}

	if _, err = notifySender.SendCustomTemplateToUser(notification_handler.NotificationChannelPush, event.UserId,
		title, body, headline, ctx); err != nil {
		return nil, err
	}

	if err = db.Create(&database.Notification{
		UserId:        event.ToUserId,
		Type:          "push.profile.following",
		Title:         title,
		Message:       body,
		RelatedUserId: null.IntFrom(event.UserId),
		CreatedAt:     time.Now().UTC(),
		ContentId:     null.IntFrom(0),
	}).Error; err != nil {
		return nil, err
	}

	if err = notification.IncrementUnreadNotificationsCounter(db, event.UserId); err != nil {
		return nil, err
	}

	return &event.Messages, nil
}
