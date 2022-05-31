package user_go

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/digitalmonsters/go-common/wrappers"
	"github.com/rs/zerolog/log"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
	"time"
)

type IUserGoWrapper interface {
	GetUsers(userIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]UserRecord]

	GetUsersDetails(userIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]UserDetailRecord]
	GetUserDetails(userId int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[UserDetailRecord]

	GetProfileBulk(currentUserId int64, userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetProfileBulkResponseChan
	GetUsersActiveThresholds(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersActiveThresholdsResponseChan
	GetUserIdsFilterByUsername(userIds []int64, searchQuery string, apmTransaction *apm.Transaction, forceLog bool) chan GetUserIdsFilterByUsernameResponseChan
	GetUsersTags(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersTagsResponseChan
	AuthGuest(deviceId string, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[AuthGuestResp]
	GetBlockList(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[string][]int64]
	GetUserBlock(blockedTo int64, blockedBy int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[UserBlockData]
	UpdateUserMetadataAfterRegistration(request UpdateUserMetaDataRequest, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[UserRecord]
	ForceResetUserWithNewGuestIdentity(deviceId string, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[ForceResetUserIdentityWithNewGuestResponse]
	VerifyUser(userId int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[UserRecord]
	GetAllActiveBots(ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[GetAllActiveBotsResponse]
	GetConfigPropertiesInternal(properties []string, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[GetConfigPropertiesResponseChan]
	UpdateEmailMarketing(userId int64, emailMarketing null.String, emailMarketingVerified bool, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[any]
	GenerateDeeplink(urlPath string, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[GenerateDeeplinkResponse]
	CreateExport(name string, exportType ExportType, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[CreateExportResponse]
	FinalizeExport(exportId int64, file null.String, err error, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[FinalizeExportResponse]
}

//goland:noinspection GoNameStartsWithPackageName
type UserGoWrapper struct {
	baseWrapper    *wrappers.BaseWrapper
	defaultTimeout time.Duration
	serviceApiUrl  string
	publicApiUrl   string
	serviceName    string
}

func NewUserGoWrapper(config boilerplate.WrapperConfig) IUserGoWrapper {
	timeout := 5 * time.Second

	if config.TimeoutSec > 0 {
		timeout = time.Duration(config.TimeoutSec) * time.Second
	}

	if len(config.ApiUrl) == 0 {
		config.ApiUrl = "http://user-go"

		log.Warn().Msgf("Api Url is missing for UserGo. Setting as default : %v", config.ApiUrl)
	}

	return &UserGoWrapper{
		baseWrapper:    wrappers.GetBaseWrapper(),
		defaultTimeout: timeout,
		serviceApiUrl:  fmt.Sprintf("%v/rpc-service", common.StripSlashFromUrl(config.ApiUrl)),
		publicApiUrl:   common.StripSlashFromUrl(config.ApiUrl),
		serviceName:    "user-go",
	}
}

func (w UserGoWrapper) GetUsers(userIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]UserRecord] {
	return wrappers.ExecuteRpcRequestAsync[map[int64]UserRecord](w.baseWrapper, w.serviceApiUrl,
		"GetUsersInternal", GetUsersRequest{
			UserIds: userIds,
		}, map[string]string{}, w.defaultTimeout, apm.TransactionFromContext(ctx), w.serviceName, forceLog)
}

func (w UserGoWrapper) GetUsersDetails(userIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]UserDetailRecord] {
	return wrappers.ExecuteRpcRequestAsync[map[int64]UserDetailRecord](w.baseWrapper, w.serviceApiUrl,
		"GetUsersDetailsInternal", GetUsersDetailRequest{
			UserIds: userIds,
		}, map[string]string{}, w.defaultTimeout, apm.TransactionFromContext(ctx), w.serviceName, forceLog)
}

func (w UserGoWrapper) GetUserDetails(userId int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[UserDetailRecord] {
	ch := make(chan wrappers.GenericResponseChan[UserDetailRecord], 2)

	go func() {
		defer func() {
			close(ch)
		}()

		resp := <-w.GetUsersDetails([]int64{userId}, ctx, forceLog)

		if resp.Error != nil {
			ch <- wrappers.GenericResponseChan[UserDetailRecord]{
				Error: resp.Error,
			}

			return
		}

		if v, ok := resp.Response[userId]; ok {
			ch <- wrappers.GenericResponseChan[UserDetailRecord]{
				Response: v,
			}
		} else {
			ch <- wrappers.GenericResponseChan[UserDetailRecord]{
				Error: &rpc.RpcError{Code: error_codes.GenericNotFoundError, Message: "item not found in dictionary"},
			}
		}
	}()

	return ch
}

func (w UserGoWrapper) GetProfileBulk(currentUserId int64, userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetProfileBulkResponseChan {
	respCh := make(chan GetProfileBulkResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.serviceApiUrl, "GetProfileBulkInternal", GetProfileBulkRequest{
		CurrentUserId: currentUserId,
		UserIds:       userIds,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	go func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := GetProfileBulkResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			data := map[int64]UserProfileDetailRecord{}

			if err := json.Unmarshal(resp.Result, &data); err != nil {
				result.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    w.baseWrapper.GetHostName(),
					ServiceName: w.serviceName,
				}
			} else {
				result.Items = data
			}
		}

		respCh <- result
	}()

	return respCh
}

func (w UserGoWrapper) GetUsersActiveThresholds(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersActiveThresholdsResponseChan {
	respCh := make(chan GetUsersActiveThresholdsResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.serviceApiUrl, "GetUsersActiveThresholds", GetUsersActiveThresholdsRequest{
		UserIds: userIds,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	go func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := GetUsersActiveThresholdsResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			var data = make(map[int64]ThresholdsStruct)

			if err := json.Unmarshal(resp.Result, &data); err != nil {
				result.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    w.baseWrapper.GetHostName(),
					ServiceName: w.serviceName,
				}
			} else {
				result.Items = data
			}
		}

		respCh <- result
	}()

	return respCh
}

func (w UserGoWrapper) GetUserIdsFilterByUsername(userIds []int64, searchQuery string, apmTransaction *apm.Transaction, forceLog bool) chan GetUserIdsFilterByUsernameResponseChan {
	respCh := make(chan GetUserIdsFilterByUsernameResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.serviceApiUrl, "GetUserIdsFilterByUsername", GetUserIdsFilterByUsernameRequest{
		UserIds:     userIds,
		SearchQuery: searchQuery,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	go func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := GetUserIdsFilterByUsernameResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			var data = make([]int64, 0)

			if err := json.Unmarshal(resp.Result, &data); err != nil {
				result.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    w.baseWrapper.GetHostName(),
					ServiceName: w.serviceName,
				}
			} else {
				result.UserIds = data
			}
		}

		respCh <- result
	}()

	return respCh
}

func (w UserGoWrapper) GetUsersTags(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersTagsResponseChan {
	respCh := make(chan GetUsersTagsResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.serviceApiUrl, "GetUsersTags", GetUsersTagsRequest{
		UserIds: userIds,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	go func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := GetUsersTagsResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			var data = make(map[int64][]Tag)

			if err := json.Unmarshal(resp.Result, &data); err != nil {
				result.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    w.baseWrapper.GetHostName(),
					ServiceName: w.serviceName,
				}
			} else {
				result.Items = data
			}
		}

		respCh <- result
	}()

	return respCh
}

func (w *UserGoWrapper) AuthGuest(deviceId string, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[AuthGuestResp] {
	return wrappers.ExecuteRpcRequestAsync[AuthGuestResp](w.baseWrapper, w.serviceApiUrl, "AuthGuestInternal", AuthGuestRequest{DeviceId: deviceId},
		map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)
}

func (w *UserGoWrapper) GetBlockList(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[string][]int64] {
	return wrappers.ExecuteRpcRequestAsync[map[string][]int64](w.baseWrapper, w.serviceApiUrl, "GetBlockListBulkInternal", GetBlockListRequest{UserIds: userIds},
		map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)
}

func (w *UserGoWrapper) GetUserBlock(blockedTo int64, blockedBy int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[UserBlockData] {
	return wrappers.ExecuteRpcRequestAsync[UserBlockData](w.baseWrapper, w.serviceApiUrl, "GetBlockListBulkInternal", GetUserBlockRequest{
		BlockBy:   blockedBy,
		BlockedTo: blockedTo,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)
}

func (w UserGoWrapper) UpdateUserMetadataAfterRegistration(request UpdateUserMetaDataRequest, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[UserRecord] {
	return wrappers.ExecuteRpcRequestAsync[UserRecord](w.baseWrapper, w.serviceApiUrl, "UpdateUserMetadataAfterRegistration", request,
		map[string]string{}, w.defaultTimeout, apm.TransactionFromContext(ctx), w.serviceName, forceLog)
}

func (w UserGoWrapper) ForceResetUserWithNewGuestIdentity(deviceId string, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[ForceResetUserIdentityWithNewGuestResponse] {
	return wrappers.ExecuteRpcRequestAsync[ForceResetUserIdentityWithNewGuestResponse](w.baseWrapper, w.serviceApiUrl,
		"ForceResetUserWithNewGuestIdentity", ForceResetUserIdentityWithNewGuestRequest{
			DeviceId: deviceId,
		}, map[string]string{}, w.defaultTimeout, apm.TransactionFromContext(ctx), w.serviceName, forceLog)
}

func (w UserGoWrapper) VerifyUser(userId int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[UserRecord] {
	return wrappers.ExecuteRpcRequestAsync[UserRecord](w.baseWrapper, w.serviceApiUrl,
		"VerifyUser", VerifyUserRequest{
			UserId: userId,
		}, map[string]string{}, w.defaultTimeout, apm.TransactionFromContext(ctx), w.serviceName, forceLog)
}

func (w UserGoWrapper) GetAllActiveBots(ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[GetAllActiveBotsResponse] {
	return wrappers.ExecuteRpcRequestAsync[GetAllActiveBotsResponse](w.baseWrapper, w.serviceApiUrl,
		"GetAllActiveBots", nil, map[string]string{}, w.defaultTimeout, apm.TransactionFromContext(ctx), w.serviceName, forceLog)
}

func (w UserGoWrapper) GetConfigPropertiesInternal(properties []string, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[GetConfigPropertiesResponseChan] {
	return wrappers.ExecuteRpcRequestAsync[GetConfigPropertiesResponseChan](w.baseWrapper, w.serviceApiUrl,
		"GetConfigPropertiesInternal", GetConfigPropertiesRequest{
			Properties: properties,
		}, map[string]string{}, w.defaultTimeout, apm.TransactionFromContext(ctx), w.serviceName, forceLog)
}

func (w UserGoWrapper) UpdateEmailMarketing(userId int64, emailMarketing null.String, emailMarketingVerified bool,
	ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[any] {
	return wrappers.ExecuteRpcRequestAsync[any](w.baseWrapper, w.serviceApiUrl,
		"UpdateEmailMarketing", UpdateEmailMarketingRequest{
			UserId:                 userId,
			EmailMarketing:         emailMarketing,
			EmailMarketingVerified: emailMarketingVerified,
		}, map[string]string{}, w.defaultTimeout, apm.TransactionFromContext(ctx), w.serviceName, forceLog)
}

func (w UserGoWrapper) GenerateDeeplink(urlPath string, ctx context.Context,
	forceLog bool) chan wrappers.GenericResponseChan[GenerateDeeplinkResponse] {
	return wrappers.ExecuteRpcRequestAsync[GenerateDeeplinkResponse](w.baseWrapper, w.serviceApiUrl,
		"GenerateDeeplink", GenerateDeeplinkRequest{
			UrlPath: urlPath,
		}, map[string]string{}, w.defaultTimeout, apm.TransactionFromContext(ctx), w.serviceName, forceLog)
}

func (w UserGoWrapper) CreateExport(name string, exportType ExportType, ctx context.Context,
	forceLog bool) chan wrappers.GenericResponseChan[CreateExportResponse] {
	return wrappers.ExecuteRpcRequestAsync[CreateExportResponse](w.baseWrapper, w.serviceApiUrl,
		"CreateExport", CreateExportRequest{
			Name: name,
			Type: exportType,
		}, map[string]string{}, w.defaultTimeout, apm.TransactionFromContext(ctx), w.serviceName, forceLog)
}

func (w UserGoWrapper) FinalizeExport(exportId int64, file null.String, err error, ctx context.Context,
	forceLog bool) chan wrappers.GenericResponseChan[FinalizeExportResponse] {
	return wrappers.ExecuteRpcRequestAsync[FinalizeExportResponse](w.baseWrapper, w.serviceApiUrl,
		"FinalizeExport", FinalizeExportRequest{
			ExportId: exportId,
			File:     file,
			Error:    err,
		}, map[string]string{}, w.defaultTimeout, apm.TransactionFromContext(ctx), w.serviceName, forceLog)
}
