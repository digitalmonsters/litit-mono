package content

import (
	"context"
	"fmt"
	"hash/fnv"
	"log"

	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/translation"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/database/scylla"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/digitalmonsters/notification-handler/pkg/utils"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
)

func process(event newSendingEvent, ctx context.Context, notifySender sender.ISender,
	contentWrapper content.IContentWrapper) (*kafka.Message, error) {
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
			var uuIds []uuid.UUID
			if err := database.GetDbWithContext(database.DbTypeReadonly, ctx).
				Table("notifications").Where("content_id = ?", event.Id).Pluck("id", &uuIds).Error; err != nil {
				apm_helper.LogError(err, ctx)
			}

			for _, pack := range lo.Chunk(uuIds, 50) {
				if err := database.GetDb(database.DbTypeMaster).
					Exec("delete from notifications where id in ?", pack).Error; err != nil {
					apm_helper.LogError(err, ctx)
				}
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

	notificationContent := &database.NotificationContent{
		Id:      event.Id,
		Width:   event.Width,
		Height:  event.Height,
		VideoId: event.VideoId,
	}

	renderData, authorLanguage, err := utils.GetUserRenderingVariablesWithLanguage(event.UserId, ctx)
	if err != nil {
		return &event.Messages, err
	}

	log.Println(event)
	if event.CrudOperation == eventsourcing.ChangeEventTypeCreated && !event.Unlisted && !event.Draft && !event.Deleted {
		if event.ContentType == eventsourcing.ContentTypeVideo {
			templateName = "content_upload"
			notificationType = "push.content.successful-upload"
		} else if event.ContentType == eventsourcing.ContentTypeSpot {
			templateName = "spot_upload"
			notificationType = "push.spot.successful-upload"
		} else if event.ContentType == eventsourcing.ContentTypeCatsSpot {
			templateName = "spot_upload_cat"
			notificationType = "push.spot_cat.successful-upload"
		} else if event.ContentType == eventsourcing.ContentTypeDogsSpot {
			templateName = "spot_upload_dog"
			notificationType = "push.spot_dog.successful-upload"
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

			if rejectReasonText == "Boring Spot" {
				return nil, nil
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

	return sendPushToFollowers(event, notificationContent, renderData, ctx, notifySender)
}

func sendPushToFollowers(event newSendingEvent, notificationContent *database.NotificationContent,
	renderingVariables database.RenderingVariables, ctx context.Context, notifySender sender.ISender) (*kafka.Message, error) {
	limit := 1000
	var pageState []byte
	session := database.GetScyllaSession()

	for {
		iter := session.Query("select entity_id, event_applied from notification_relation where user_id = ? and event_type = 'follow'",
			event.UserId).WithContext(ctx).PageSize(limit).PageState(pageState).Iter()

		pageState = iter.PageState()
		scanner := iter.Scanner()

		var followersIds []int64
		for scanner.Next() {
			var entityId int64
			var eventApplied bool

			if err := scanner.Scan(&entityId, &eventApplied); err != nil {
				return nil, errors.WithStack(err)
			}

			if eventApplied {
				followersIds = append(followersIds, entityId)
			}
		}

		if err := scanner.Err(); err != nil {
			return nil, errors.WithStack(err)
		}
		if err := iter.Close(); err != nil {
			return nil, errors.WithStack(err)
		}

		for _, followerId := range followersIds {
			var language translation.Language

			userIter := session.Query("select language from user where cluster_key = ? and user_id = ?",
				scylla.GetUserClusterKey(followerId), followerId).Iter()
			userIter.Scan(&language)

			if err := userIter.Close(); err != nil {
				return &event.Messages, errors.WithStack(err)
			}

			language = translation.DefaultUserLanguage

			var shouldRetry bool

			h := fnv.New32a()
			_, err := h.Write([]byte(fmt.Sprintf("%v%v", event.Id, event.UserId)))
			if err != nil {
				return &event.Messages, errors.WithStack(err)
			}

			entityId := int64(h.Sum32())

			shouldRetry, err = notifySender.PushNotification(database.Notification{
				UserId:             followerId,
				Type:               "push.content.new-posted",
				RelatedUserId:      null.IntFrom(event.UserId),
				ContentId:          null.IntFrom(event.Id),
				Content:            notificationContent,
				RenderingVariables: renderingVariables,
			}, entityId, 0, "content_posted", language, "default", ctx)
			if err != nil {
				if shouldRetry {
					return nil, errors.WithStack(err)
				}

				return &event.Messages, errors.WithStack(err)
			}
		}

		if len(followersIds) < limit {
			break
		}
	}

	return &event.Messages, nil
}
