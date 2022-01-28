package category

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
	SendSmsInternal(message string, phoneNumber string, apmTransaction *apm.Transaction, forceLog bool) chan SendMessageResponseChan
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
		serviceName:    "notification-gateway",
	}
}

func (w *Wrapper) SendSmsInternal(message string, phoneNumber string, apmTransaction *apm.Transaction, forceLog bool) chan SendMessageResponseChan {
	respCh := make(chan SendMessageResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "SendSmsInternal", SendMessageRequest{
		Message:     message,
		PhoneNumber: phoneNumber,
	}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	go func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := SendMessageResponseChan{
			Error: resp.Error,
		}
		respCh <- result
	}()

	return respCh
}
