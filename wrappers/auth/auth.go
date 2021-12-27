package auth

import (
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/digitalmonsters/go-common/wrappers"
	"go.elastic.co/apm"
	"time"
)

type IAuthWrapper interface {
	ParseToken(token string, ignoreExpiration bool, apmTransaction *apm.Transaction,
		forceLog bool) chan AuthParseTokenResponseChan
	GenerateToken(userId int64, apmTransaction *apm.Transaction,
		forceLog bool) chan GenerateTokenResponseChan
}

type AuthWrapper struct {
	defaultTimeout time.Duration
	apiUrl         string
	serviceName    string
	baseWrapper    *wrappers.BaseWrapper
}

func NewAuthWrapper(config boilerplate.WrapperConfig) IAuthWrapper {
	timeout := 5 * time.Second

	if config.TimeoutSec > 0 {
		timeout = time.Duration(config.TimeoutSec) * time.Second
	}

	return &AuthWrapper{
		defaultTimeout: timeout,
		apiUrl:         common.StripSlashFromUrl(config.ApiUrl),
		serviceName:    "forward-auth",
		baseWrapper:    wrappers.GetBaseWrapper(),
	}
}

func (w *AuthWrapper) ParseToken(token string, ignoreExpiration bool, apmTransaction *apm.Transaction,
	forceLog bool) chan AuthParseTokenResponseChan {
	resChan := make(chan AuthParseTokenResponseChan, 2)

	w.baseWrapper.GetPool().Submit(func() {
		rpcInternalResponse := <-w.baseWrapper.SendRequestWithRpcResponse(fmt.Sprintf("%v/token/parse", w.apiUrl),
			"unpack jwt",
			AuthParseTokenRequest{
				Token:            token,
				IgnoreExpiration: ignoreExpiration,
			}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

		finalResponse := AuthParseTokenResponseChan{
			Error: rpcInternalResponse.Error,
		}

		if len(rpcInternalResponse.Result) > 0 {
			if err := json.Unmarshal(rpcInternalResponse.Result, &finalResponse.Resp); err != nil {
				finalResponse.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    w.baseWrapper.GetHostName(),
					ServiceName: w.serviceName,
				}
			}
		}

		resChan <- finalResponse
	})

	return resChan
}

func (w *AuthWrapper) GenerateToken(userId int64, apmTransaction *apm.Transaction,
	forceLog bool) chan GenerateTokenResponseChan {
	resChan := make(chan GenerateTokenResponseChan, 2)

	w.baseWrapper.GetPool().Submit(func() {
		rpcInternalResponse := <-w.baseWrapper.SendRequestWithRpcResponseFromAnyService(fmt.Sprintf("%v/token/%v", w.apiUrl, userId),
			"GET",
			"application/json",
			"generate token",
			nil, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

		finalResponse := GenerateTokenResponseChan{
			Error: rpcInternalResponse.Error,
		}

		if len(rpcInternalResponse.Result) > 0 {
			if err := json.Unmarshal(rpcInternalResponse.Result, &finalResponse.Resp); err != nil {
				finalResponse.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    w.baseWrapper.GetHostName(),
					ServiceName: w.serviceName,
				}
			}
		}

		resChan <- finalResponse
	})

	return resChan
}
