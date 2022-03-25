package user_delete

import (
	"context"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/segmentio/kafka-go"
)

func process(event newSendingEvent, ctx context.Context) (*kafka.Message, error) {
	if event.CrudOperation != eventsourcing.ChangeEventTypeDeleted {
		return &event.Messages, nil
	}

	db := database.GetDb(database.DbTypeMaster).WithContext(ctx)
	tx := db.Begin()
	defer tx.Rollback()

	var userIds []int64
	if err := tx.Model(&database.Notification{}).Where("related_user_id = ?", event.UserId).Select("user_id").Find(&userIds).Error; err != nil {
		return nil, err
	}

	if len(userIds) > 0 {
		if err := tx.Exec("delete from notifications where related_user_id = ?", event.UserId).Error; err != nil {
			return nil, err
		}

		if err := tx.Exec("update user_notifications set unread_count = unread_count - 1 where user_id in ?", userIds).Error; err != nil {
			return nil, err
		}
	}

	if err := tx.Exec("delete from notifications where user_id = ?", event.UserId).Error; err != nil {
		return nil, err
	}

	if err := tx.Exec("delete from user_notifications where user_id = ?", event.UserId).Error; err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &event.Messages, nil
}
