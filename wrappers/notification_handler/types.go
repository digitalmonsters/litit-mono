package notification_handler

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/wrappers"
	"time"
)

type INotificationHandlerWrapper interface {
	EnqueueNotificationWithTemplate(templateName string, userId int64,
		renderingVars map[string]string, ctx context.Context) chan EnqueueMessageResult
}

//goland:noinspection GoNameStartsWithPackageName
type NotificationHandlerWrapper struct {
	baseWrapper    *wrappers.BaseWrapper
	defaultTimeout time.Duration
	apiUrl         string
	serviceName    string
	publisher      eventsourcing.IEventPublisher
}

type EnqueueMessageResult struct {
	Id    string `json:"id"`
	Error error  `json:"error"`
}

type SendNotificationWithTemplate struct {
	Id                 string            `json:"id"`
	TemplateName       string            `json:"template_name"`
	RenderingVariables map[string]string `json:"rendering_variables"`
	UserId             int64             `json:"user_id"`
}

func (s SendNotificationWithTemplate) GetPublishKey() string {
	return fmt.Sprint(s.UserId)
}
