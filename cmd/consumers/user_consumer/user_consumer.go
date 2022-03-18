package user_consumer

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/comments/pkg/database"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
)

func process(event wrappedEvent, apmTransaction *apm.Transaction, ctx context.Context) (*kafka.Message, error) {
	apm_helper.AddApmLabel(apmTransaction, "user_id", event.UserId)
	apm_helper.AddApmLabel(apmTransaction, "crud_operation_reason", event.BaseChangeEvent.CrudOperationReason)
	apm_helper.AddApmLabel(apmTransaction, "crud_operation", event.BaseChangeEvent.CrudOperation)

	if event.BaseChangeEvent.CrudOperation != eventsourcing.ChangeEventTypeDeleted &&
		event.BaseChangeEvent.CrudOperationReason != eventsourcing.DeleteModeHard {

		return &event.Message, nil
	}

	tx := database.GetDb().WithContext(ctx).Begin()
	defer tx.Rollback()

	var votes []database.CommentVote

	if err := tx.Where("vote_up is not null").
		Where("user_id = ?", event.UserId).Find(&votes).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	for _, v := range votes {
		queryPart := "num_upvotes = num_upvotes - 1"

		if !v.VoteUp.Bool {
			queryPart = "num_downvotes = num_downvotes - 1"
		}

		if err := tx.Exec(fmt.Sprintf("update comment set %v where id = ?",
			queryPart), v.CommentId).Error; err != nil {
			return nil, errors.WithStack(err)
		}
	}

	var toDelete []database.Comment // two levels only supported

	if err := tx.Where("author_id = ?", event.UserId).Find(&toDelete).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	for _, c := range toDelete {
		if err := tx.Exec("delete from comment where id = ?", c.Id).Error; err != nil {
			return nil, errors.WithStack(err)
		}

		if c.ParentId.Valid {
			if err := tx.Exec("update comment set num_replies = num_replies - 1 where id = ?",
				c.ParentId.ValueOrZero()).Error; err != nil {
				return nil, errors.WithStack(err)
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return &event.Message, nil
}
