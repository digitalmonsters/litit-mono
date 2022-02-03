package notification_gateway

import (
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/wrappers"
	"go.elastic.co/apm"
	"time"
)

type Wrapper struct {
	baseWrapper    *wrappers.BaseWrapper
	defaultTimeout time.Duration
	apiUrl         string
	serviceName    string
}

type INotificationGatewayWrapper interface {
	SendSmsInternal(message string, phoneNumber string, apmTransaction *apm.Transaction, forceLog bool) chan SendSmsMessageResponseChan
	SendEmailInternal(ccAddresses, toAddresses []string, htmlBody, textBody, subject string, apmTransaction *apm.Transaction, forceLog bool) chan SendEmailMessageResponseChan
}

func NewNotificationGatewayWrapper(config boilerplate.WrapperConfig) INotificationGatewayWrapper {
	timeout := 5 * time.Second

	if config.TimeoutSec > 0 {
		timeout = time.Duration(config.TimeoutSec) * time.Second
	}

	return &Wrapper{
		baseWrapper:    wrappers.GetBaseWrapper(),
		defaultTimeout: timeout,
		apiUrl:         fmt.Sprintf("%v/rpc-service", common.StripSlashFromUrl(config.ApiUrl)),
		serviceName:    "notification_gateway",
	}
}

func (w *Wrapper) SendSmsInternal(message string, phoneNumber string, apmTransaction *apm.Transaction, forceLog bool) chan SendSmsMessageResponseChan {
	respCh := make(chan SendSmsMessageResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "SendSmsInternal", SendSmsMessageRequest{
		Message:     message,
		PhoneNumber: phoneNumber,
	}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	go func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := SendSmsMessageResponseChan{
			Error: resp.Error,
		}
		respCh <- result
	}()

	return respCh
}

func (w *Wrapper) SendEmailInternal(ccAddresses, toAddresses []string, htmlBody, textBody, subject string, apmTransaction *apm.Transaction, forceLog bool) chan SendEmailMessageResponseChan {
	respCh := make(chan SendEmailMessageResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "SendEmailInternal", SendEmailMessageRequest{
		CcAddresses: ccAddresses,
		ToAddresses: toAddresses,
		HtmlBody:    htmlBody,
		TextBody:    textBody,
		Subject:     subject,
	}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	go func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := SendEmailMessageResponseChan{
			Error: resp.Error,
		}
		respCh <- result
	}()

	return respCh
}
