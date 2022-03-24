package auth_go

import (
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/digitalmonsters/go-common/wrappers"
	"github.com/rs/zerolog/log"
	"go.elastic.co/apm"
	"time"
)

type IAuthGoWrapper interface {
	CheckAdminPermissions(userId int64, obj string, transaction *apm.Transaction, forceLog bool) chan CheckAdminPermissionsResponseChan
	CheckLegacyAdmin(userId int64, transaction *apm.Transaction, forceLog bool) chan CheckLegacyAdminResponseChan
	GetAdminIdsFilterByEmail(adminIds []int64, searchQuery string, apmTransaction *apm.Transaction, forceLog bool) chan GetAdminIdsFilterByEmailResponseChan
	GetAdminsInfoById(adminIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetAdminsInfoByIdResponseChan
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

	if len(config.ApiUrl) == 0 {
		config.ApiUrl = "http://auth-go"

		log.Warn().Msgf("Api Url is missing for AuthGo. Setting as default : %v", config.ApiUrl)
	}

	return &AuthGoWrapper{
		defaultTimeout: timeout,
		apiUrl:         fmt.Sprintf("%v/rpc-service", common.StripSlashFromUrl(config.ApiUrl)),
		serviceName:    "auth_go",
		baseWrapper:    wrappers.GetBaseWrapper(),
	}
}

func (w AuthGoWrapper) AddNewUser(req eventsourcing.UserEvent, apmTransaction *apm.Transaction, forceLog bool) chan AddUserResponseChan {
	respCh := make(chan AddUserResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "AddNewUser",
		req, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	go func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := AddUserResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			data := eventsourcing.UserEvent{} // no need ot have it

			if err := json.Unmarshal(resp.Result, &data); err != nil {
				result.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    w.baseWrapper.GetHostName(),
					ServiceName: w.serviceName,
				}
			} else {
				result.Item = data
			}
		}

		respCh <- result
	}()

	return respCh
}

func (w *AuthGoWrapper) CheckLegacyAdmin(userId int64, transaction *apm.Transaction, forceLog bool) chan CheckLegacyAdminResponseChan {
	respCh := make(chan CheckLegacyAdminResponseChan, 2)

	rpcInternalResponseCh := w.baseWrapper.SendRpcRequest(w.apiUrl, "CheckLegacyAdmin", CheckLegacyAdminRequest{
		UserId: userId,
	}, map[string]string{}, w.defaultTimeout, transaction, w.serviceName, forceLog)

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

func (w *AuthGoWrapper) CheckAdminPermissions(userId int64, obj string, transaction *apm.Transaction,
	forceLog bool) chan CheckAdminPermissionsResponseChan {
	respCh := make(chan CheckAdminPermissionsResponseChan, 2)

	rpcInternalResponseCh := w.baseWrapper.SendRpcRequest(w.apiUrl, "CheckUserAdminPermissions", CheckAdminPermissionsRequest{
		UserId: userId,
		Object: obj,
	}, map[string]string{}, w.defaultTimeout, transaction, w.serviceName, forceLog)

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

func (u AuthGoWrapper) GetAdminIdsFilterByEmail(adminIds []int64, searchQuery string, apmTransaction *apm.Transaction, forceLog bool) chan GetAdminIdsFilterByEmailResponseChan {
	respCh := make(chan GetAdminIdsFilterByEmailResponseChan, 2)

	respChan := u.baseWrapper.SendRpcRequest(u.apiUrl, "GetAdminIdsFilterByEmail", GetAdminIdsFilterByEmailRequest{
		AdminIds:    adminIds,
		SearchQuery: searchQuery,
	}, map[string]string{}, u.defaultTimeout, apmTransaction, u.serviceName, forceLog)

	go func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := GetAdminIdsFilterByEmailResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			var data = make([]int64, 0)

			if err := json.Unmarshal(resp.Result, &data); err != nil {
				result.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    u.baseWrapper.GetHostName(),
					ServiceName: u.serviceName,
				}
			} else {
				result.AdminIds = data
			}
		}

		respCh <- result
	}()

	return respCh
}

func (u AuthGoWrapper) GetAdminsInfoById(adminIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetAdminsInfoByIdResponseChan {
	respCh := make(chan GetAdminsInfoByIdResponseChan, 2)

	respChan := u.baseWrapper.SendRpcRequest(u.apiUrl, "GetAdminsInfoById", GetAdminsInfoByIdRequest{
		AdminIds: adminIds,
	}, map[string]string{}, u.defaultTimeout, apmTransaction, u.serviceName, forceLog)

	go func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := GetAdminsInfoByIdResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			var data = make(map[int64]AdminGeneralInfo)

			if err := json.Unmarshal(resp.Result, &data); err != nil {
				result.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    u.baseWrapper.GetHostName(),
					ServiceName: u.serviceName,
				}
			} else {
				result.Items = data
			}
		}

		respCh <- result
	}()

	return respCh
}
