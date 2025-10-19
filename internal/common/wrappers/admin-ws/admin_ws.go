package admin_ws

import (
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/digitalmonsters/go-common/wrappers"
	"github.com/rs/zerolog/log"
	"go.elastic.co/apm"
	"time"
)

type IAdminWsWrapper interface {
	SendMessage(event EventType, message interface{}, transaction *apm.Transaction, forceLog bool) chan SendMessageResponseCh
}

type AdminWsWrapper struct {
	defaultTimeout time.Duration
	apiUrl         string
	serviceName    string
	baseWrapper    *wrappers.BaseWrapper
}

func NewAdminWsWrapper(config boilerplate.WrapperConfig) IAdminWsWrapper {
	timeout := 5 * time.Second

	if config.TimeoutSec > 0 {
		timeout = time.Duration(config.TimeoutSec) * time.Second
	}

	if len(config.ApiUrl) == 0 {
		config.ApiUrl = "http://admin-ws"

		log.Warn().Msgf("Api Url is missing for AuthGo. Setting as default : %v", config.ApiUrl)
	}

	return &AdminWsWrapper{
		defaultTimeout: timeout,
		apiUrl:         fmt.Sprintf("%v/rpc-service", common.StripSlashFromUrl(config.ApiUrl)),
		serviceName:    "admin_ws",
		baseWrapper:    wrappers.GetBaseWrapper(),
	}
}

func (w *AdminWsWrapper) SendMessage(event EventType, message interface{}, transaction *apm.Transaction,
	forceLog bool) chan SendMessageResponseCh {
	respCh := make(chan SendMessageResponseCh, 2)

	messageMarshaled, err := json.Marshal(message)
	if err != nil {
		go func() {
			respCh <- SendMessageResponseCh{
				Error: &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Hostname:    w.apiUrl,
					ServiceName: w.serviceName,
				},
			}
		}()

		return respCh
	}

	rpcInternalResponseCh := w.baseWrapper.SendRpcRequest(w.apiUrl, "SendMessage", SendMessageRequest{
		Event:   event,
		Message: messageMarshaled,
	}, map[string]string{}, w.defaultTimeout, transaction, w.serviceName, forceLog)

	go func() {
		resp := <-rpcInternalResponseCh

		respCh <- SendMessageResponseCh{
			Error: resp.Error,
		}
	}()

	return respCh
}
