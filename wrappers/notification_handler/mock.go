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
	CreateNotificationFn                    func(notifications Notification, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[CreateNotificationResponse]
	DeleteNotificationByIntroIDFn           func(introID int, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[DeleteNotificationByIntroIDResponse]
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

func (m *NotificationHandlerWrapperMock) CreateNotification(notifications Notification, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[CreateNotificationResponse] {
	return m.CreateNotificationFn(notifications, ctx, forceLog)
}

func (m *NotificationHandlerWrapperMock) DeleteNotificationByIntroID(introID int, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[DeleteNotificationByIntroIDResponse] {
	return m.DeleteNotificationByIntroIDFn(introID, ctx, forceLog)
}

func (m *NotificationHandlerWrapperMock) SendGenericEmail(To, Subject, Body string, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[GenericEmailResponse] {
	return m.SendGenericEmail(To, Subject, Body, ctx, forceLog)
}
func (m *NotificationHandlerWrapperMock) SendGenericHTMLEmail(To, Subject, Body string, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[GenericEmailResponse] {
	return m.SendGenericHTMLEmail(To, Subject, Body, ctx, forceLog)
}

func GetMock() INotificationHandlerWrapper { // for compiler errors
	return &NotificationHandlerWrapperMock{}
}
