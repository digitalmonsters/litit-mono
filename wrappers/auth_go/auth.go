package auth_go

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

type IAuthGoWrapper interface {
	CheckAdminPermissions(userId int64, obj string, action common.AccessLevel, transaction *apm.Transaction, forceLog bool) chan CheckAdminPermissionsResponseChan
	CheckLegacyAdmin(userId int64, transaction *apm.Transaction, forceLog bool) chan CheckLegacyAdminResponseChan
}

type AuthGoWrapper struct {
	defaultTimeout time.Duration
	apiUrl         string
	serviceName    string
	baseWrapper    *wrappers.BaseWrapper
}

func NewAuthGoWrapper(config boilerplate.WrapperConfig) IAuthGoWrapper {
	timeout := 5 * time.Second

	if config.TimeoutSec > 0 {
		timeout = time.Duration(config.TimeoutSec) * time.Second
	}

	return &AuthGoWrapper{
		defaultTimeout: timeout,
		apiUrl:         fmt.Sprintf("%v/rpc-service", common.StripSlashFromUrl(config.ApiUrl)),
		serviceName:    "auth_go",
		baseWrapper:    wrappers.GetBaseWrapper(),
	}
}

func (w *AuthGoWrapper) CheckLegacyAdmin(userId int64, transaction *apm.Transaction, forceLog bool) chan CheckLegacyAdminResponseChan {
	respCh := make(chan CheckLegacyAdminResponseChan, 2)

	rpcInternalResponseCh := w.baseWrapper.SendRpcRequest(w.apiUrl, "CheckLegacyAdmin", CheckLegacyAdminRequest{
		UserId: userId,
	}, w.defaultTimeout, transaction, w.serviceName, forceLog)

	go func() {
		resp := <-rpcInternalResponseCh

		finalResponse := CheckLegacyAdminResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			if err := json.Unmarshal(resp.Result, &finalResponse.Resp); err != nil {
				finalResponse.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    w.baseWrapper.GetHostName(),
					ServiceName: w.serviceName,
				}
			}
		}
		respCh <- finalResponse
	}()

	return respCh
}

func (w *AuthGoWrapper) CheckAdminPermissions(userId int64, obj string, action common.AccessLevel, transaction *apm.Transaction,
	forceLog bool) chan CheckAdminPermissionsResponseChan {
	respCh := make(chan CheckAdminPermissionsResponseChan, 2)

	rpcInternalResponseCh := w.baseWrapper.SendRpcRequest(w.apiUrl, "CheckUserAdminPermissions", CheckAdminPermissionsRequest{
		UserId:      userId,
		Method:      obj,
		AccessLevel: action,
	}, w.defaultTimeout, transaction, w.serviceName, forceLog)

	go func() {

		resp := <-rpcInternalResponseCh

		finalResponse := CheckAdminPermissionsResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			if err := json.Unmarshal(resp.Result, &finalResponse.Resp); err != nil {
				finalResponse.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    w.baseWrapper.GetHostName(),
					ServiceName: w.serviceName,
				}
			}
		}

		respCh <- finalResponse
	}()

	return respCh
}
