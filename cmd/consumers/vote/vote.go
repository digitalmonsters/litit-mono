package vote

import (
	"context"
	"github.com/digitalmonsters/go-common/eventsourcing"
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
	"strconv"
	"time"
)

func process(event newSendingEvent, ctx context.Context, notifySender sender.ISender, userGoWrapper user_go.IUserGoWrapper,
	apmTransaction *apm.Transaction) (*kafka.Message, error) {
	if !event.Upvote.Valid {
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

	renderData := map[string]string{
		"firstname": userData.Firstname,
		"lastname":  userData.Lastname,
		"comment":   event.Comment,
	}

	db := database.GetDb(database.DbTypeMaster).WithContext(ctx)

	var title string
	var body string
	var headline string
	var templateName string

	if event.Upvote.Bool {
		templateName = "comment_vote_like"
	} else {
		templateName = "comment_vote_dislike"
	}

	title, body, headline, _, err = notifySender.RenderTemplate(db, templateName, renderData)
	if err == renderer.TemplateRenderingError {
		return &event.Messages, err // we should continue, no need to retry
	} else if err != nil {
		return nil, err
	}

	if _, err = notifySender.SendCustomTemplateToUser(notification_handler.NotificationChannelPush, event.CommentAuthorId,
		title, body, headline, ctx); err != nil {
		return nil, err
	}

	notificationContent := database.NotificationComment{
		Id:       event.CommentId,
		Type:     database.NotificationCommentTypeContent,
		Comment:  event.Comment,
		ParentId: event.ParentId,
		Upvote:   event.Upvote,
	}

	reason, _ := strconv.Atoi(event.CrudOperationReason)
	commentChangeReason := eventsourcing.CommentChangeReason(reason)

	if commentChangeReason == eventsourcing.CommentChangeReasonContent {
		notificationContent.ContentId = null.IntFrom(event.EntityId)
	} else if commentChangeReason == eventsourcing.CommentChangeReasonProfile {
		notificationContent.ProfileId = null.IntFrom(event.EntityId)
	}

	if err = db.Create(&database.Notification{
		UserId:        event.CommentAuthorId,
		Type:          "push.comment.vote",
		Title:         title,
		Message:       body,
		CreatedAt:     time.Now().UTC(),
		CommentId:     null.IntFrom(event.CommentId),
		Comment:       &notificationContent,
		RelatedUserId: null.IntFrom(event.UserId),
	}).Error; err != nil {
		return nil, err
	}

	if err = notificationPkg.IncrementUnreadNotificationsCounter(db, event.CommentAuthorId); err != nil {
		return nil, err
	}

	return &event.Messages, nil
}
