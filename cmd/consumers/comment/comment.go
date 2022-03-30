package comment

import (
	"context"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/wrappers/comment"
	"github.com/digitalmonsters/go-common/wrappers/content"
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
	"strconv"
	"time"
)

func process(event newSendingEvent, ctx context.Context, notifySender sender.ISender, userGoWrapper user_go.IUserGoWrapper,
	contentWrapper content.IContentWrapper, commentWrapper comment.ICommentWrapper, apmTransaction *apm.Transaction) (*kafka.Message, error) {
	if event.CrudOperation != eventsourcing.ChangeEventTypeCreated {
		return &event.Messages, nil
	}

	var userData user_go.UserRecord
	var err error

	resp := <-userGoWrapper.GetUsers([]int64{event.AuthorId}, apmTransaction, false)
	if resp.Error != nil {
		return nil, resp.Error.ToError()
	}

	var ok bool
	if userData, ok = resp.Items[event.AuthorId]; !ok {
		return &event.Messages, errors.WithStack(errors.New("user not found")) // we should continue, no need to retry
	}

	renderData := map[string]string{
		"firstname": userData.Firstname,
		"lastname":  userData.Lastname,
		"comment":   event.Comment.Comment,
	}

	db := database.GetDb(database.DbTypeMaster).WithContext(ctx)

	var title string
	var body string
	var headline string

	notificationComment := &database.NotificationComment{
		Id:        event.Id,
		Comment:   event.Comment.Comment,
		ParentId:  event.ParentId,
		ContentId: event.ContentId,
		ProfileId: event.ProfileId,
	}
	var notificationContent *database.NotificationContent

	var contentAuthorId null.Int
	if event.ContentId.Valid {
		contentResp := <-contentWrapper.GetInternal([]int64{event.ContentId.Int64}, false, apmTransaction, false)
		if contentResp.Error != nil {
			return nil, err
		}

		notificationContent = &database.NotificationContent{Id: event.ContentId.Int64}

		if simpleContent, ok := contentResp.Items[event.ContentId.Int64]; ok {
			notificationContent.Width = simpleContent.Width
			notificationContent.Height = simpleContent.Height
			notificationContent.VideoId = simpleContent.VideoId
			contentAuthorId = null.IntFrom(simpleContent.AuthorId)
		}
	}

	commentResp := <-commentWrapper.GetCommentsInfoById([]int64{event.Id}, apmTransaction, false)
	if commentResp.Error != nil {
		return nil, err
	}

	var parentAuthorId null.Int

	if commentInfo, ok := commentResp.Items[event.Id]; ok {
		parentAuthorId = commentInfo.ParentAuthorId
	}

	if parentAuthorId.Valid && parentAuthorId.Int64 != event.AuthorId && contentAuthorId.Valid {
		title, body, headline, _, err = notifySender.RenderTemplate(db, "comment_reply", renderData)
		if err == renderer.TemplateRenderingError {
			return &event.Messages, err // we should continue, no need to retry
		} else if err != nil {
			return nil, err
		}

		if _, err = notifySender.SendCustomTemplateToUser(notification_handler.NotificationChannelPush, contentAuthorId.Int64,
			title, body, headline, ctx); err != nil {
			return nil, err
		}

		notificationComment.Type = database.NotificationCommentTypeContent

		if err = db.Create(&database.Notification{
			UserId:        parentAuthorId.Int64,
			Type:          "push.comment.reply",
			Title:         title,
			Message:       body,
			RelatedUserId: null.IntFrom(event.AuthorId),
			CommentId:     null.IntFrom(event.Id),
			Comment:       notificationComment,
			ContentId:     event.ContentId,
			Content:       notificationContent,
			CreatedAt:     time.Now().UTC(),
		}).Error; err != nil {
			return nil, err
		}

		if err = notification.IncrementUnreadNotificationsCounter(db, contentAuthorId.Int64); err != nil {
			return nil, err
		}
	}

	reason, _ := strconv.Atoi(event.CrudOperationReason)

	switch eventsourcing.CommentChangeReason(reason) {
	case eventsourcing.CommentChangeReasonContent:
		if !contentAuthorId.Valid || contentAuthorId.Int64 == event.AuthorId {
			return &event.Messages, nil
		}

		title, body, headline, _, err = notifySender.RenderTemplate(db, "comment_content_resource_create", renderData)
		if err == renderer.TemplateRenderingError {
			return &event.Messages, err // we should continue, no need to retry
		} else if err != nil {
			return nil, err
		}

		if _, err = notifySender.SendCustomTemplateToUser(notification_handler.NotificationChannelPush, contentAuthorId.Int64,
			title, body, headline, ctx); err != nil {
			return nil, err
		}

		notificationComment.Type = database.NotificationCommentTypeContent

		if err = db.Create(&database.Notification{
			UserId:        contentAuthorId.Int64,
			Type:          "push.content.comment",
			Title:         title,
			Message:       body,
			RelatedUserId: null.IntFrom(event.AuthorId),
			CommentId:     null.IntFrom(event.Id),
			Comment:       notificationComment,
			ContentId:     event.ContentId,
			Content:       notificationContent,
			CreatedAt:     time.Now().UTC(),
		}).Error; err != nil {
			return nil, err
		}

		if err = notification.IncrementUnreadNotificationsCounter(db, contentAuthorId.Int64); err != nil {
			return nil, err
		}
	case eventsourcing.CommentChangeReasonProfile:
		if event.ProfileId.Int64 == event.AuthorId {
			return &event.Messages, nil
		}

		title, body, headline, _, err = notifySender.RenderTemplate(db, "comment_profile_resource_create", renderData)
		if err == renderer.TemplateRenderingError {
			return &event.Messages, err // we should continue, no need to retry
		} else if err != nil {
			return nil, err
		}

		if _, err = notifySender.SendCustomTemplateToUser(notification_handler.NotificationChannelPush, event.ProfileId.Int64,
			title, body, headline, ctx); err != nil {
			return nil, err
		}

		notificationComment.Type = database.NotificationCommentTypeProfile

		if err = db.Create(&database.Notification{
			UserId:        event.ProfileId.Int64,
			Type:          "push.profile.comment",
			Title:         title,
			Message:       body,
			RelatedUserId: null.IntFrom(event.AuthorId),
			CommentId:     null.IntFrom(event.Id),
			Comment:       notificationComment,
			ContentId:     event.ContentId,
			Content:       notificationContent,
			CreatedAt:     time.Now().UTC(),
		}).Error; err != nil {
			return nil, err
		}

		if err = notification.IncrementUnreadNotificationsCounter(db, event.ProfileId.Int64); err != nil {
			return nil, err
		}
	}

	return &event.Messages, nil
}
