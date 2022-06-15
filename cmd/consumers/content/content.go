package content

import (
	"context"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/translation"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/digitalmonsters/go-common/wrappers/follow"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
)

func process(event newSendingEvent, ctx context.Context, notifySender sender.ISender, followWrapper follow.IFollowWrapper,
	userGoWrapper user_go.IUserGoWrapper, contentWrapper content.IContentWrapper) (*kafka.Message, error) {
	db := database.GetDb(database.DbTypeMaster).WithContext(ctx)

	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "user_id", event.UserId)
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "author_id", event.UserId)
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "content_id", event.Id)
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "crud_operation_reason", event.BaseChangeEvent.CrudOperationReason)
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "crud_operation", event.BaseChangeEvent.CrudOperation)

	if event.CrudOperation == eventsourcing.ChangeEventTypeDeleted {
		tx := db.Begin()
		defer tx.Rollback()

		var userIds []int64
		if err := tx.Model(&database.Notification{}).Where("content_id = ?", event.Id).Select("user_id").Find(&userIds).Error; err != nil {
			return nil, err
		}

		if len(userIds) > 0 {
			if err := tx.Exec("delete from notifications where content_id = ?", event.Id).Error; err != nil {
				return nil, err
			}

			if err := tx.Exec("update user_notifications set unread_count = unread_count - 1 where user_id in ?", userIds).Error; err != nil {
				return nil, err
			}
		}

		if err := tx.Commit().Error; err != nil {
			return nil, err
		}

		return &event.Messages, nil
	}

	var err error
	var notificationType string
	var templateName string
	renderData := map[string]string{}

	notificationContent := &database.NotificationContent{
		Id:      event.Id,
		Width:   event.Width,
		Height:  event.Height,
		VideoId: event.VideoId,
	}

	var authorLanguage translation.Language

	resp := <-userGoWrapper.GetUsers([]int64{event.UserId}, ctx, false)
	if resp.Error != nil {
		return nil, resp.Error.ToError()
	}

	if userData, ok := resp.Response[event.UserId]; !ok {
		return &event.Messages, errors.WithStack(errors.New("user not found")) // we should continue, no need to retry
	} else {
		authorLanguage = userData.Language
	}

	if event.CrudOperation == eventsourcing.ChangeEventTypeCreated && !event.Unlisted && !event.Draft && !event.Deleted {
		if event.ContentType == eventsourcing.ContentTypeVideo {
			templateName = "content_upload"
			notificationType = "push.content.successful-upload"
		} else if event.ContentType == eventsourcing.ContentTypeSpot {
			templateName = "spot_upload"
			notificationType = "push.spot.successful-upload"
		}
	} else if event.CrudOperation == eventsourcing.ChangeEventTypeUpdated {
		if string(event.CrudOperationReason) == "rejected" {
			rejectReasonText := ""

			if event.RejectReason.Valid {
				rejectReasonResp := <-contentWrapper.GetRejectReason([]int64{event.RejectReason.Int64}, true, ctx, false)
				if rejectReasonResp.Error != nil {
					return nil, err
				}

				if rejectReason, ok := rejectReasonResp.Response[event.RejectReason.Int64]; ok {
					rejectReasonText = rejectReason.Reason
				} else {
					rejectReasonText = "unknown reason"
				}
			}

			renderData = map[string]string{
				"reason": rejectReasonText,
			}
			templateName = "content_reject"
			notificationType = "push.content.rejected"
		} else if event.IsNewVisible {
			templateName = "content_upload"
			notificationType = "push.content.successful-upload"
		} else {
			return &event.Messages, nil
		}
	} else {
		return &event.Messages, nil
	}

	shouldRetry, err := notifySender.PushNotification(database.Notification{
		UserId:             event.UserId,
		Type:               notificationType,
		RelatedUserId:      null.IntFrom(event.UserId),
		ContentId:          null.IntFrom(event.Id),
		Content:            notificationContent,
		RenderingVariables: renderData,
	}, event.Id, event.UserId, templateName, authorLanguage, "default", ctx)
	if err != nil {
		if shouldRetry {
			return nil, errors.WithStack(err)
		}

		return &event.Messages, err
	}

	if templateName != "content_upload" {
		return &event.Messages, nil
	}

	followersCountResp := <-followWrapper.GetFollowersCount([]int64{event.UserId}, apm.TransactionFromContext(ctx), false)
	if followersCountResp.Error != nil {
		return nil, followersCountResp.Error.ToError()
	}

	if len(followersCountResp.Data) == 0 {
		return &event.Messages, nil
	}

	var ok bool

	limit, ok := followersCountResp.Data[event.UserId]
	if !ok {
		return &event.Messages, nil
	}

	userFollowersResp := <-followWrapper.GetUserFollowers(event.UserId, "", int(limit), apm.TransactionFromContext(ctx), false)
	if userFollowersResp.Error != nil {
		return nil, userFollowersResp.Error.ToError()
	}

	if len(userFollowersResp.FollowerIds) == 0 {
		return &event.Messages, nil
	}

	//var userData user_go.UserRecord
	//
	//resp = <-userGoWrapper.GetUsers([]int64{event.UserId}, ctx, false)
	//if resp.Error != nil {
	//	return nil, resp.Error.ToError()
	//}
	//
	//if userData, ok = resp.Response[event.UserId]; !ok {
	//	return &event.Messages, errors.WithStack(errors.New("user not found")) // we should continue, no need to retry
	//}

	//templateName = "content_posted"
	//notificationType = "push.content.new-posted"

	//firstName, lastName := userData.GetFirstAndLastNameWithPrivacy()

	//renderData = map[string]string{
	//	"firstname": firstName,
	//	"lastname":  lastName,
	//}

	//title, body, headline, _, err = notifySender.RenderTemplate(tx, templateName, renderData, userData.Language)
	//if err == renderer.TemplateRenderingError {
	//	return &event.Messages, err // we should continue, no need to retry
	//} else if err != nil {
	//	return nil, err
	//}

	//for _, followerId := range userFollowersResp.FollowerIds {
	//	if _, err = notifySender.SendCustomTemplateToUser(notification_handler.NotificationChannelPush, followerId, templateName, "default",
	//		title, body, headline, nil, ctx); err != nil {
	//		return nil, err
	//	}
	//
	//	if err = tx.Create(&database.Notification{
	//		UserId:             followerId,
	//		Type:               notificationType,
	//		Title:              title,
	//		Message:            body,
	//		ContentId:          null.IntFrom(event.Id),
	//		Content:            notificationContent,
	//		CreatedAt:          time.Now().UTC(),
	//		RelatedUserId:      null.IntFrom(event.UserId),
	//		RenderingVariables: renderData,
	//	}).Error; err != nil {
	//		return nil, err
	//	}
	//
	//	if err = notification.IncrementUnreadNotificationsCounter(tx, followerId); err != nil {
	//		return nil, err
	//	}
	//}

	//if err = tx.Commit().Error; err != nil {
	//	return nil, err
	//}

	return &event.Messages, nil
}
