package sender

import (
	"context"
	"github.com/digitalmonsters/go-common/wrappers/notification_gateway"
	"github.com/digitalmonsters/go-common/wrappers/notification_handler"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"gorm.io/gorm"
)

type ISender interface {
	SendTemplateToUser(channel notification_handler.NotificationChannel,
		templateName string, userId int64, renderingData map[string]string,
		ctx context.Context) (interface{}, error)

	SendCustomTemplateToUser(channel notification_handler.NotificationChannel, userId int64, title, body, headline string, ctx context.Context) (interface{}, error)
	RenderTemplate(db *gorm.DB, templateName string,
		renderingData map[string]string) (title string, body string, headline string, renderingTemplate database.RenderTemplate, err error)
	SendEmail(msg []notification_gateway.SendEmailMessageRequest, ctx context.Context) error
}
