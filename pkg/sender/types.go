package sender

import (
	"context"

	"github.com/digitalmonsters/go-common/translation"
	"github.com/digitalmonsters/go-common/wrappers/notification_gateway"
	"github.com/digitalmonsters/notification-handler/pkg/database"
)

type ISender interface {
	SendEmail(msg []notification_gateway.SendEmailMessageRequest, ctx context.Context) error
	PushNotification(notification database.Notification, imageUrl string, entityId int64, relatedEntityId int64,
		templateName string, language translation.Language, customKind string, ctx context.Context) (shouldRetry bool, innerErr error)
	UnapplyEvent(userId int64, eventType string, entityId int64, relatedEntityId int64, ctx context.Context) error
}
