package notification_gateway

import (
	"go.elastic.co/apm"
)

type NotificationGatewayWrapperMock struct {
	SendSmsInternalFn   func(message string, phoneNumber string, apmTransaction *apm.Transaction, forceLog bool) chan SendSmsMessageResponseChan
	SendEmailInternalFn func(ccAddresses, toAddresses []string, htmlBody, textBody, subject string, apmTransaction *apm.Transaction, forceLog bool) chan SendEmailMessageResponseChan
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
