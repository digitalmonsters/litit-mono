package auth

import (
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/valyala/fasthttp"
	"go.elastic.co/apm"
	"time"
)

type IAuthWrapper interface {
}

type AuthWrapper struct {
	defaultTimeout time.Duration
	apiUrl         string
	serviceName    string
	client         *fasthttp.Client
}

func NewAuthWrapper(apiUrl string) IAuthWrapper {
	return &AuthWrapper{defaultTimeout: 5 * time.Second, apiUrl: apiUrl,
		serviceName: "auth-wrapper", client: &fasthttp.Client{}}
}

func (w *AuthWrapper) ParseToken2(token string, ignoreExpiration bool, apmTransaction *apm.Transaction,
	forceLog bool) chan AuthParseTokenResponseChan {
	resChan := make(chan AuthParseTokenResponseChan, 2)

	go func() {
		var chanResponse AuthParseTokenResponseChan

		defer func() {
			resChan <- chanResponse
		}()

		req := fasthttp.AcquireRequest()
		resp := fasthttp.AcquireResponse()

		defer func() {
			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
		}()

		req.Header.SetMethod("POST")

		if err := json.NewEncoder(resp.BodyWriter()).Encode(AuthParseTokenRequest{
			Token:            token,
			IgnoreExpiration: ignoreExpiration,
		}); err != nil {
			chanResponse.Error = &rpc.RpcError{
				Code:    error_codes.GenericMappingError,
				Message: err.Error(),
				Data:    nil,
				Stack:   fmt.Sprintf("%+v", err),
			}

			return
		}

		if err := apm_helper.SendHttpRequest(w.client, req, resp, apmTransaction, w.defaultTimeout, forceLog); err != nil {
			chanResponse.Error = &rpc.RpcError{
				Code:    error_codes.GenericServerError,
				Message: err.Error(),
				Data:    nil,
				Stack:   fmt.Sprintf("%+v", err),
			}

			return
		}

		var result rpc.RpcResponseInternal

		if err := json.Unmarshal(resp.Body(), &result); err != nil {
			chanResponse.Error = &rpc.RpcError{
				Code:    error_codes.GenericMappingError,
				Message: err.Error(),
				Data:    nil,
				Stack:   fmt.Sprintf("%+v", err),
			}

			return
		}

		var realResponse AuthParseTokenResponse

		if result.Error != nil {
			chanResponse.Error = result.Error

			return
		}

		if len(result.Result) > 0 {
			if err := json.Unmarshal(result.Result, &realResponse); err != nil {
				chanResponse.Error = &rpc.RpcError{
					Code:    error_codes.GenericMappingError,
					Message: err.Error(),
					Data:    nil,
					Stack:   fmt.Sprintf("%+v", err),
				}

				return
			}
		}

		chanResponse.Resp = realResponse
	}()

	return resChan
}

type AuthParseTokenRequest struct {
	Token            string `json:"token"`
	IgnoreExpiration bool   `json:"ignore_expiration"`
}

type AuthParseTokenResponseChan struct {
	Resp  AuthParseTokenResponse
	Error *rpc.RpcError
}

type AuthParseTokenResponse struct {
	UserId  int64 `json:"user_id"`
	Expired bool  `json:"expired"`
}
