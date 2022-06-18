package vote

import (
	"context"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"gopkg.in/guregu/null.v4"
	"strconv"
)

func process(event newSendingEvent, ctx context.Context, notifySender sender.ISender,
	userGoWrapper user_go.IUserGoWrapper) (*kafka.Message, error) {
	if !event.Upvote.Valid || event.CommentAuthorId == event.UserId {
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

	renderData := database.RenderingVariables{
		"firstname": firstName,
		"lastname":  lastName,
		"comment":   event.Comment,
	}

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

	shouldRetry, err := notifySender.PushNotification(database.Notification{
		UserId:             event.CommentAuthorId,
		Type:               "push.comment.vote",
		CommentId:          null.IntFrom(event.CommentId),
		Comment:            &notificationContent,
		RelatedUserId:      null.IntFrom(event.UserId),
		RenderingVariables: renderData,
	}, event.CommentId, event.UserId, templateName, userData.Language, "default", ctx)
	if err != nil {
		if shouldRetry {
			return nil, errors.WithStack(err)
		}

		return &event.Messages, err
	}

	return &event.Messages, nil
}
