package content

import (
	"context"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/frontend"
	"github.com/digitalmonsters/go-common/wrappers/notification_handler"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/notification"
	"github.com/digitalmonsters/notification-handler/pkg/renderer"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/segmentio/kafka-go"
	"gopkg.in/guregu/null.v4"
	"time"
)

func process(event newSendingEvent, ctx context.Context, notifySender sender.ISender) (*kafka.Message, error) {
	db := database.GetDb(database.DbTypeMaster).WithContext(ctx)

	if event.CrudOperation == eventsourcing.ChangeEventTypeDeleted {
		tx := db.Begin()
		defer tx.Rollback()

		var userIds []int64
		if err := tx.Model(&database.Notification{}).Where("content_id = ?", event.Id).Select("user_id").Row().Scan(&userIds); err != nil {
			return nil, err
		}

		if len(userIds) > 0 {
			if err := tx.Model(&database.Notification{}).Delete("content_id = ?", event.Id).Error; err != nil {
				return nil, err
			}

			if err := tx.Exec("update set unread_count = unread_count - 1 where user_id in ?", userIds).Error; err != nil {
				return nil, err
			}
		}

		if err := tx.Commit().Error; err != nil {
			return nil, err
		}

		return &event.Messages, nil
	}

	var err error
	var title string
	var body string
	var headline string
	var notificationType string
	var templateName string
	renderData := map[string]string{}

	notificationContent := &database.NotificationContent{
		Id:      event.Id,
		Width:   event.Width,
		Height:  event.Height,
		VideoId: event.VideoId,
	}

	if event.CrudOperation == eventsourcing.ChangeEventTypeCreated && !event.Unlisted && !event.Draft && !event.Deleted {
		templateName = "content_upload"
		notificationType = "push.content.successful-upload"
	} else if event.CrudOperation == eventsourcing.ChangeEventTypeUpdated {
		if string(event.CrudOperationReason) == "rejected" {
			renderData = map[string]string{
				"reason": getRejectionReason(event.RejectReason),
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

	title, body, headline, _, err = notifySender.RenderTemplate(db, templateName, renderData)
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
		UserId:    event.UserId,
		Type:      notificationType,
		Title:     title,
		Message:   body,
		ContentId: null.IntFrom(event.Id),
		Content:   notificationContent,
		CreatedAt: time.Now().UTC(),
	}).Error; err != nil {
		return nil, err
	}

	if err = notification.IncrementUnreadNotificationsCounter(db, event.UserId); err != nil {
		return nil, err
	}

	return &event.Messages, nil
}

func getRejectionReason(reason frontend.RejectReason) string {
	reasonStr := ""

	switch reason {
	case frontend.RejectReasonFakeIdentity:
		reasonStr = "fake identity"
	case frontend.RejectReasonOffensive:
		reasonStr = "offensive"
	case frontend.RejectReasonHateSpeech:
		reasonStr = "hate speech"
	case frontend.RejectReasonHateNudityOrSexualActivity:
		reasonStr = "hate nudity or sexual activity"
	case frontend.RejectReasonViolence:
		reasonStr = "violence"
	case frontend.RejectReasonHarassment:
		reasonStr = "harassment"
	case frontend.RejectReasonNone:
	default:
		reasonStr = "no reason"
	}

	return reasonStr
}
