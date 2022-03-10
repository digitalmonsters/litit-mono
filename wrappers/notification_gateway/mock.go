package notification_gateway

import (
	"context"
	"go.elastic.co/apm"
)

type NotificationGatewayWrapperMock struct {
	SendSmsInternalFn    func(message string, phoneNumber string, apmTransaction *apm.Transaction, forceLog bool) chan SendSmsMessageResponseChan
	SendEmailInternalFn  func(ccAddresses, toAddresses []string, htmlBody, textBody, subject string, apmTransaction *apm.Transaction, forceLog bool) chan SendEmailMessageResponseChan
	EnqueuePushForUserFn func(tokens []string, deviceType DeviceType, title string, body string, extraData map[string]string,
		userId int64, ctx context.Context) chan error
}

func (w *NotificationGatewayWrapperMock) EnqueuePushForUser(tokens []string, deviceType DeviceType, title string, body string, extraData map[string]string,
	userId int64, ctx context.Context) chan error {
	return w.EnqueuePushForUserFn(tokens, deviceType, title, body, extraData, userId, ctx)
}

func (w *NotificationGatewayWrapperMock) SendSmsInternal(message string, phoneNumber string, apmTransaction *apm.Transaction, forceLog bool) chan SendSmsMessageResponseChan {
	return w.SendSmsInternalFn(message, phoneNumber, apmTransaction, forceLog)
}

func (w *NotificationGatewayWrapperMock) SendEmailInternal(ccAddresses, toAddresses []string, htmlBody, textBody, subject string, apmTransaction *apm.Transaction, forceLog bool) chan SendEmailMessageResponseChan {
	return w.SendEmailInternalFn(ccAddresses, toAddresses, htmlBody, textBody, subject, apmTransaction, forceLog)
}

func GetMock() INotificationGatewayWrapper {
	return &NotificationGatewayWrapperMock{}
}
