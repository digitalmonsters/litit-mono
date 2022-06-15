package sender

import (
	"context"
	"github.com/digitalmonsters/go-common/translation"
	"github.com/digitalmonsters/go-common/wrappers/notification_gateway"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"gorm.io/gorm"
)

type ISender interface {
	RenderTemplate(db *gorm.DB, templateName string, renderingData map[string]string,
		language translation.Language) (title string, body string, headline string, titleMultiple string, bodyMultiple string,
		headlineMultiple string, renderingTemplate database.RenderTemplate, err error)
	SendEmail(msg []notification_gateway.SendEmailMessageRequest, ctx context.Context) error
	PushNotification(notification database.Notification, entityId int64, relatedEntityId int64,
		templateName string, language translation.Language, customKind string, ctx context.Context) (shouldRetry bool, innerErr error)
}
