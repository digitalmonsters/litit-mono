package user_delete

import (
	"context"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
)

func process(event newSendingEvent, ctx context.Context, apmTransaction *apm.Transaction) (*kafka.Message, error) {
	if event.CrudOperation != eventsourcing.ChangeEventTypeDeleted {
		return &event.Messages, nil
	}

	apm_helper.AddApmLabel(apmTransaction, "crud_operation_reason", event.BaseChangeEvent.CrudOperationReason)
	apm_helper.AddApmLabel(apmTransaction, "crud_operation", event.BaseChangeEvent.CrudOperation)
	apm_helper.AddApmLabel(apmTransaction, "user_id", event.UserId)

	db := database.GetDb(database.DbTypeMaster).WithContext(ctx)
	tx := db.Begin()
	defer tx.Rollback()

	var userIds []int64
	if err := tx.Model(&database.Notification{}).Where("related_user_id = ?", event.UserId).Select("user_id").Find(&userIds).Error; err != nil {
		return nil, err
	}

	if len(userIds) > 0 {
		var uuIds []uuid.UUID
		if err := database.GetDbWithContext(database.DbTypeReadonly, ctx).
			Table("notifications").Where("related_user_id = ?", event.UserId).Pluck("id", &uuIds).Error; err != nil {
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

	var uuIds []uuid.UUID
	if err := database.GetDbWithContext(database.DbTypeReadonly, ctx).
		Table("notifications").Where("user_id = ?", event.UserId).Pluck("id", &uuIds).Error; err != nil {
		apm_helper.LogError(err, ctx)
	}

	for _, pack := range lo.Chunk(uuIds, 50) {
		if err := database.GetDb(database.DbTypeMaster).
			Exec("delete from notifications where id in ?", pack).Error; err != nil {
			apm_helper.LogError(err, ctx)
		}
	}

	if err := tx.Exec("delete from user_notifications where user_id = ?", event.UserId).Error; err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &event.Messages, nil
}
