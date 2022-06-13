package sender

import (
	"context"
	"github.com/digitalmonsters/go-common/translation"
	"github.com/digitalmonsters/go-common/wrappers/notification_gateway"
	"github.com/digitalmonsters/go-common/wrappers/notification_handler"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"gorm.io/gorm"
)

type ISender interface {
	SendTemplateToUser(channel notification_handler.NotificationChannel,
		title, body, headline string, renderingTemplate database.RenderTemplate, userId int64, renderingData map[string]string,
		customData database.CustomData, isGrouped bool, ctx context.Context) (interface{}, error)

	SendCustomTemplateToUser(channel notification_handler.NotificationChannel, userId int64, pushType, kind,
		title, body, headline string, customData database.CustomData, isGrouped bool, ctx context.Context) (interface{}, error)
	RenderTemplate(db *gorm.DB, templateName string, renderingData map[string]string,
		language translation.Language) (title string, body string, headline string, renderingTemplate database.RenderTemplate, err error)
	SendEmail(msg []notification_gateway.SendEmailMessageRequest, ctx context.Context) error
}
