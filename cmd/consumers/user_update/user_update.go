package user_update

import (
	"context"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/database/scylla"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
)

func process(event newSendingEvent, ctx context.Context) (*kafka.Message, error) {
	if event.CrudOperation != eventsourcing.ChangeEventTypeUpdated && event.CrudOperationReason != "general_info_updated" {
		return &event.Messages, nil
	}

	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "crud_operation_reason", event.BaseChangeEvent.CrudOperationReason)
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "crud_operation", event.BaseChangeEvent.CrudOperation)
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "user_id", event.UserId)

	session := database.GetScyllaSession()

	if err := session.Query("update user set username = ?, firstname = ?, lastname = ?, name_privacy_status = ?, "+
		"language = ?, email = ? where cluster_key = ? and user_id = ?", event.Username.ValueOrZero(),
		event.Firstname.ValueOrZero(), event.Lastname.ValueOrZero(), event.NamePrivacyStatus, event.Language,
		event.Email.ValueOrZero(), scylla.GetUserClusterKey(event.UserId), event.UserId).WithContext(ctx).Exec(); err != nil {
		return nil, errors.WithStack(err)
	}

	return &event.Messages, nil
}
