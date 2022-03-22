package sender

import "context"

type NotificationChannel byte

const (
	NotificationChannelPush = NotificationChannel(1)
)

type ISender interface {
	SendTemplateToUser(channel NotificationChannel,
		templateName string, userId int64, renderingData map[string]string,
		ctx context.Context) (interface{}, error)

	SendCustomTemplateToUser(channel NotificationChannel, userId int64, title, body, headline string, ctx context.Context) (interface{}, error)
}
