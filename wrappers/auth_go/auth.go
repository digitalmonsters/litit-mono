package auth_go

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/digitalmonsters/go-common/wrappers"
	"github.com/rs/zerolog/log"
	"go.elastic.co/apm"
)

type IAuthGoWrapper interface {
	CheckAdminPermissions(userId int64, obj string, transaction *apm.Transaction, forceLog bool) chan CheckAdminPermissionsResponseChan
	CheckLegacyAdmin(userId int64, transaction *apm.Transaction, forceLog bool) chan CheckLegacyAdminResponseChan
	GetAdminIdsFilterByEmail(adminIds []int64, searchQuery string, apmTransaction *apm.Transaction, forceLog bool) chan GetAdminIdsFilterByEmailResponseChan
	GetAdminsInfoById(adminIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetAdminsInfoByIdResponseChan
	AddNewUser(req eventsourcing.UserEvent, apmTransaction *apm.Transaction, forceLog bool) chan AddUserResponseChan
	IsGuest(userId int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[IsGuestResponse]
	GetUsersRegistrationType(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[int64]SocialProviderType]
	InternalGetUsersForValidation(userIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]UserForValidator]
	UpdateEmailForUser(userId int64, email string, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]UpdateEmailForUserResponse]
	GetOnlineUsers(forceLog bool) chan wrappers.GenericResponseChan[OnlineUserResponse]
	TriggerUserOnline(userId int64) chan wrappers.GenericResponseChan[GenericTriggerOnlineOfflineRequest]
	TriggerUserOffline(userId int64) chan wrappers.GenericResponseChan[GenericTriggerOnlineOfflineRequest]
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

func (u AuthGoWrapper) IsGuest(userId int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[IsGuestResponse] {
	return wrappers.ExecuteRpcRequestAsync[IsGuestResponse](u.baseWrapper, u.apiUrl, "IsGuest", IsGuestRequest{
		UserId: userId,
	}, map[string]string{}, u.defaultTimeout, apmTransaction, u.serviceName, forceLog)
}

func (u AuthGoWrapper) UpdateEmailForUser(userId int64, email string, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]UpdateEmailForUserResponse] {
	return wrappers.ExecuteRpcRequestAsync[map[int64]UpdateEmailForUserResponse](u.baseWrapper,
		u.apiUrl, "UpdateEmailForUser", UpdateEmailForUserRequest{
			UserId:  userId,
			EmailId: email,
		}, map[string]string{}, 5*time.Second, apm.TransactionFromContext(ctx), u.serviceName, forceLog)
}

func (u AuthGoWrapper) AddNewUser(req eventsourcing.UserEvent, apmTransaction *apm.Transaction, forceLog bool) chan AddUserResponseChan {
	respCh := make(chan AddUserResponseChan, 2)

	respChan := u.baseWrapper.SendRpcRequest(u.apiUrl, "AddNewUser",
		req, map[string]string{}, u.defaultTimeout, apmTransaction, u.serviceName, forceLog)

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
					Hostname:    u.baseWrapper.GetHostName(),
					ServiceName: u.serviceName,
				}
			} else {
				result.Item = data
			}
		}

		respCh <- result
	}()

	return respCh
}

func (u *AuthGoWrapper) CheckLegacyAdmin(userId int64, transaction *apm.Transaction, forceLog bool) chan CheckLegacyAdminResponseChan {
	respCh := make(chan CheckLegacyAdminResponseChan, 2)

	rpcInternalResponseCh := u.baseWrapper.SendRpcRequest(u.apiUrl, "CheckLegacyAdmin", CheckLegacyAdminRequest{
		UserId: userId,
	}, map[string]string{}, u.defaultTimeout, transaction, u.serviceName, forceLog)

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
					Hostname:    u.baseWrapper.GetHostName(),
					ServiceName: u.serviceName,
				}
			}
		}
		respCh <- finalResponse
	}()

	return respCh
}

func (u *AuthGoWrapper) CheckAdminPermissions(userId int64, obj string, transaction *apm.Transaction,
	forceLog bool) chan CheckAdminPermissionsResponseChan {
	respCh := make(chan CheckAdminPermissionsResponseChan, 2)

	rpcInternalResponseCh := u.baseWrapper.SendRpcRequest(u.apiUrl, "CheckUserAdminPermissions", CheckAdminPermissionsRequest{
		UserId: userId,
		Object: obj,
	}, map[string]string{}, u.defaultTimeout, transaction, u.serviceName, forceLog)

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
					Hostname:    u.baseWrapper.GetHostName(),
					ServiceName: u.serviceName,
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

func (u AuthGoWrapper) GetUsersRegistrationType(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[int64]SocialProviderType] {
	return wrappers.ExecuteRpcRequestAsync[map[int64]SocialProviderType](u.baseWrapper, u.apiUrl, "GetUsersRegistrationType", GetUsersRegistrationTypeRequest{
		UserIds: userIds,
	}, map[string]string{}, u.defaultTimeout, apmTransaction, u.serviceName, forceLog)
}

func (u AuthGoWrapper) GetOnlineUsers(forceLog bool) chan wrappers.GenericResponseChan[OnlineUserResponse] {
	return wrappers.ExecuteRpcRequestAsync[OnlineUserResponse](u.baseWrapper, u.apiUrl, "GetOnlineUsers", nil, map[string]string{}, u.defaultTimeout, nil, u.serviceName, forceLog)
}

func (u AuthGoWrapper) TriggerUserOnline(userId int64) chan wrappers.GenericResponseChan[GenericTriggerOnlineOfflineRequest] {
	return wrappers.ExecuteRpcRequestAsync[GenericTriggerOnlineOfflineRequest](u.baseWrapper, u.apiUrl, "TriggerUserOnline", GenericTriggerOnlineOfflineRequest{
		UserId: userId,
	}, map[string]string{}, u.defaultTimeout, nil, u.serviceName, false)
}

func (u AuthGoWrapper) TriggerUserOffline(userId int64) chan wrappers.GenericResponseChan[GenericTriggerOnlineOfflineRequest] {
	return wrappers.ExecuteRpcRequestAsync[GenericTriggerOnlineOfflineRequest](u.baseWrapper, u.apiUrl, "TriggerUserOffline", GenericTriggerOnlineOfflineRequest{
		UserId: userId,
	}, map[string]string{}, u.defaultTimeout, nil, u.serviceName, false)
}

func (u AuthGoWrapper) InternalGetUsersForValidation(userIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]UserForValidator] {
	return wrappers.ExecuteRpcRequestAsync[map[int64]UserForValidator](u.baseWrapper,
		u.apiUrl, "InternalGetUsersForValidation", InternalGetUsersForValidatorFromCacheRequest{
			UserIds: userIds,
		}, map[string]string{}, 5*time.Second, apm.TransactionFromContext(ctx), u.serviceName, forceLog)
}
