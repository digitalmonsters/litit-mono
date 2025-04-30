package email_notification

import (
	"context"

	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/notification-handler/configs"
	"github.com/digitalmonsters/notification-handler/pkg/kafka"
	"github.com/digitalmonsters/notification-handler/pkg/mail"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
)

func InitListener(appCtx context.Context, configuration boilerplate.KafkaListenerConfiguration,
	notificationSender sender.ISender, emailLinks configs.EmailLinks, emailSvc mail.IEmailService, userGoWrapper user_go.IUserGoWrapper) {
	kafka.StartEmailConsumer(appCtx, configuration.Topic, configuration.GroupId, configuration.Hosts, emailSvc, userGoWrapper)
}
