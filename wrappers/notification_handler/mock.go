package notification_handler

import (
	"context"
	"github.com/digitalmonsters/go-common/wrappers"
)

type NotificationHandlerWrapperMock struct {
	EnqueueNotificationWithTemplateFn func(templateName string, userId int64,
		renderingVars map[string]string, customData map[string]interface{}, ctx context.Context) chan EnqueueMessageResult
	EnqueueNotificationWithCustomTemplateFn func(title, body, headline string, userId int64, customData map[string]interface{}, ctx context.Context) chan EnqueueMessageResult
	GetNotificationsReadCountFn             func(notificationIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]int64]
	DisableUnregisteredTokensFn             func(tokens []string, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[[]string]
}

func (m *NotificationHandlerWrapperMock) EnqueueNotificationWithTemplate(templateName string, userId int64,
	renderingVars map[string]string, customData map[string]interface{}, ctx context.Context) chan EnqueueMessageResult {
	return m.EnqueueNotificationWithTemplateFn(templateName, userId, renderingVars, customData, ctx)
}

func (m *NotificationHandlerWrapperMock) EnqueueNotificationWithCustomTemplate(title, body, headline string, userId int64,
	customData map[string]interface{}, ctx context.Context) chan EnqueueMessageResult {
	return m.EnqueueNotificationWithCustomTemplateFn(title, body, headline, userId, customData, ctx)
}

func (m *NotificationHandlerWrapperMock) GetNotificationsReadCount(notificationIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]int64] {
	return m.GetNotificationsReadCountFn(notificationIds, ctx, forceLog)
}

func (m *NotificationHandlerWrapperMock) DisableUnregisteredTokens(tokens []string, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[[]string] {
	return m.DisableUnregisteredTokensFn(tokens, ctx, forceLog)
}

func GetMock() INotificationHandlerWrapper { // for compiler errors
	return &NotificationHandlerWrapperMock{}
}
