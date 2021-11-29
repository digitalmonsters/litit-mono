package auth

import (
	"encoding/json"
	"fmt"
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
}

type AuthWrapper struct {
	defaultTimeout time.Duration
	apiUrl         string
	serviceName    string
	baseWrapper    *wrappers.BaseWrapper
}

func NewAuthWrapper(apiUrl string) IAuthWrapper {
	return &AuthWrapper{defaultTimeout: 5 * time.Second, apiUrl: common.StripSlashFromUrl(apiUrl),
		serviceName: "auth-wrapper", baseWrapper: wrappers.GetBaseWrapper()}
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
					Code:    error_codes.GenericMappingError,
					Message: err.Error(),
					Data:    nil,
				}
			}
		}

		resChan <- finalResponse
	})

	return resChan
}
