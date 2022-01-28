package notification_gateway

import (
	"go.elastic.co/apm"
)

type NotificationGatewayWrapperMock struct {
	SendSmsInternalFn func(message string, phoneNumber string, apmTransaction *apm.Transaction, forceLog bool) chan SendMessageResponseChan
}

func (w *NotificationGatewayWrapperMock) SendSmsInternal(message string, phoneNumber string, apmTransaction *apm.Transaction, forceLog bool) chan SendMessageResponseChan {
	return w.SendSmsInternalFn(message, phoneNumber, apmTransaction, forceLog)
}

func GetMock() INotificationGatewayWrapper {
	return &NotificationGatewayWrapperMock{}
}
