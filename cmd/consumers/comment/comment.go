package comment

import (
	"context"
	"strconv"

	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/wrappers/comment"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/digitalmonsters/notification-handler/pkg/utils"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
)

func process(event newSendingEvent, ctx context.Context, notifySender sender.ISender, contentWrapper content.IContentWrapper,
	commentWrapper comment.ICommentWrapper) (*kafka.Message, error) {
	if event.CrudOperation != eventsourcing.ChangeEventTypeCreated {
		return &event.Messages, nil
	}

	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "crud_operation_reason", event.BaseChangeEvent.CrudOperationReason)
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "crud_operation", event.BaseChangeEvent.CrudOperation)
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "user_id", event.AuthorId)
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "profile_id", event.ProfileId.ValueOrZero())
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "content_id", event.ContentId.ValueOrZero())

	var err error

	notificationComment := &database.NotificationComment{
		Id:        event.Id,
		Comment:   event.Comment.Comment,
		ParentId:  event.ParentId,
		ContentId: event.ContentId,
		ProfileId: event.ProfileId,
	}
	var notificationContent *database.NotificationContent

	targetEntity := event.ProfileId.ValueOrZero()

	var contentAuthorId null.Int
	if event.ContentId.Valid {
		contentResp := <-contentWrapper.GetInternal([]int64{event.ContentId.Int64}, false, apm.TransactionFromContext(ctx), false)
		if contentResp.Error != nil {
			return nil, err
		}

		notificationContent = &database.NotificationContent{Id: event.ContentId.Int64}

		if simpleContent, ok := contentResp.Response[event.ContentId.Int64]; ok {
			notificationContent.Width = simpleContent.Width
			notificationContent.Height = simpleContent.Height
			notificationContent.VideoId = simpleContent.VideoId
			contentAuthorId = null.IntFrom(simpleContent.AuthorId)
			targetEntity = contentAuthorId.ValueOrZero()
		}
	}

	commentResp := <-commentWrapper.GetCommentsInfoById([]int64{event.Id}, apm.TransactionFromContext(ctx), false)
	if commentResp.Error != nil {
		return nil, err
	}

	var parentAuthorId null.Int

	if commentInfo, ok := commentResp.Items[event.Id]; ok {
		parentAuthorId = commentInfo.ParentAuthorId
	}

	renderData, language, err := utils.GetUserRenderingVariablesWithLanguage(targetEntity, ctx)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	renderData["comment"] = event.Comment.Comment

	reason, _ := strconv.Atoi(event.CrudOperationReason)
	commentChangeReason := eventsourcing.CommentChangeReason(reason)

	if parentAuthorId.Valid && parentAuthorId.Int64 != event.AuthorId {
		var templateName = "comment_reply"

		if commentChangeReason == eventsourcing.CommentChangeReasonContent {
			notificationComment.Type = database.NotificationCommentTypeContent
		} else {
			notificationComment.Type = database.NotificationCommentTypeProfile
		}

		shouldRetry, err := notifySender.PushNotification(database.Notification{
			UserId:             parentAuthorId.Int64,
			Type:               "push.comment.reply",
			RelatedUserId:      null.IntFrom(event.AuthorId),
			CommentId:          null.IntFrom(event.Id),
			Comment:            notificationComment,
			ContentId:          event.ContentId,
			Content:            notificationContent,
			RenderingVariables: renderData,
		}, "", notificationComment.ParentId.ValueOrZero(), event.AuthorId, templateName, language, "default", ctx)
		if err != nil {
			if shouldRetry {
				return nil, errors.WithStack(err)
			}

			return &event.Messages, err
		}

		if (contentAuthorId.Valid && parentAuthorId.Int64 == contentAuthorId.Int64) || (event.ProfileId.Valid && parentAuthorId.Int64 == event.ProfileId.Int64) {
			return &event.Messages, nil
		}
	}

	switch commentChangeReason {
	case eventsourcing.CommentChangeReasonContent:
		if !contentAuthorId.Valid || contentAuthorId.Int64 == event.AuthorId {
			return &event.Messages, nil
		}

		var templateName = "comment_content_resource_create"

		notificationComment.Type = database.NotificationCommentTypeContent

		shouldRetry, err := notifySender.PushNotification(database.Notification{
			UserId:             contentAuthorId.Int64,
			Type:               "push.content.comment",
			RelatedUserId:      null.IntFrom(event.AuthorId),
			CommentId:          null.IntFrom(event.Id),
			Comment:            notificationComment,
			ContentId:          event.ContentId,
			Content:            notificationContent,
			RenderingVariables: renderData,
		}, event.ContentId.ValueOrZero(), event.AuthorId, templateName, language, "default", ctx)
		if err != nil {
			if shouldRetry {
				return nil, errors.WithStack(err)
			}

			return &event.Messages, err
		}
	case eventsourcing.CommentChangeReasonProfile:
		if event.ProfileId.Int64 == event.AuthorId {
			return &event.Messages, nil
		}

		var templateName = "comment_profile_resource_create"

		notificationComment.Type = database.NotificationCommentTypeProfile

		var renderDataAuthor database.RenderingVariables
		renderDataAuthor, _, err = utils.GetUserRenderingVariablesWithLanguage(event.AuthorId, ctx)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		renderDataAuthor["comment"] = event.Comment.Comment

		shouldRetry, err := notifySender.PushNotification(database.Notification{
			UserId:             event.ProfileId.Int64,
			Type:               "push.profile.comment",
			RelatedUserId:      null.IntFrom(event.AuthorId),
			CommentId:          null.IntFrom(event.Id),
			Comment:            notificationComment,
			ContentId:          event.ContentId,
			Content:            notificationContent,
			RenderingVariables: renderDataAuthor,
		}, event.ProfileId.Int64, 0, templateName, language, "default", ctx)
		if err != nil {
			if shouldRetry {
				return nil, errors.WithStack(err)
			}

			return &event.Messages, err
		}
	}

	return &event.Messages, nil
}
