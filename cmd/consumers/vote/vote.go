package vote

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/digitalmonsters/notification-handler/pkg/utils"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"gopkg.in/guregu/null.v4"
	"hash/fnv"
	"strconv"
)

func process(event newSendingEvent, ctx context.Context, notifySender sender.ISender) (*kafka.Message, error) {
	if !event.Upvote.Valid || event.CommentAuthorId == event.UserId {
		return &event.Messages, nil
	}

	var err error

	renderData, language, err := utils.GetUserRenderingVariablesWithLanguage(event.UserId, ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	renderData["comment"] = event.Comment

	var templateName string

	if event.Upvote.Bool {
		templateName = "comment_vote_like"
	} else {
		templateName = "comment_vote_dislike"
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

	h := fnv.New32a()
	_, err = h.Write([]byte(fmt.Sprintf("%v%v", event.CommentId, event.UserId)))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	entityId := int64(h.Sum32())

	shouldRetry, err := notifySender.PushNotification(database.Notification{
		UserId:             event.CommentAuthorId,
		Type:               "push.comment.vote",
		CommentId:          null.IntFrom(event.CommentId),
		Comment:            &notificationContent,
		RelatedUserId:      null.IntFrom(event.UserId),
		RenderingVariables: renderData,
	}, entityId, 0, templateName, language, "default", ctx)
	if err != nil {
		if shouldRetry {
			return nil, errors.WithStack(err)
		}

		return &event.Messages, err
	}

	return &event.Messages, nil
}
