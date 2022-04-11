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
	"github.com/patrickmn/go-cache"
	"github.com/rs/zerolog/log"
	"go.elastic.co/apm"
	"time"
)

type IUserGoWrapper interface {
	GetUsers(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersResponseChan

	GetUsersDetails(userIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]UserDetailRecord]
	GetUserDetails(userId int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[UserDetailRecord]

	GetProfileBulk(currentUserId int64, userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetProfileBulkResponseChan
	GetUsersActiveThresholds(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersActiveThresholdsResponseChan
	GetUserIdsFilterByUsername(userIds []int64, searchQuery string, apmTransaction *apm.Transaction, forceLog bool) chan GetUserIdsFilterByUsernameResponseChan
	GetUsersTags(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersTagsResponseChan
	AuthGuest(deviceId string, apmTransaction *apm.Transaction, forceLog bool) chan AuthGuestResponseChan
	GetBlockList(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetBlockListResponseChan
	GetUserBlock(blockedTo int64, blockedBy int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUserBlockResponseChan
	UpdateUserMetadataAfterRegistration(request UpdateUserMetaDataRequest, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[UserRecord]
	ForceResetUserWithNewGuestIdentity(deviceId string, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[ForceResetUserIdentityWithNewGuestResponse]
	VerifyUser(userId int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[UserRecord]
}

//goland:noinspection GoNameStartsWithPackageName
type UserGoWrapper struct {
	baseWrapper    *wrappers.BaseWrapper
	defaultTimeout time.Duration
	serviceApiUrl  string
	publicApiUrl   string
	serviceName    string
	cache          *cache.Cache
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
		cache:          cache.New(4*time.Minute, 5*time.Minute),
	}
}

func (w UserGoWrapper) GetUsers(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersResponseChan {
	respCh := make(chan GetUsersResponseChan, 2)

	cachedItems := map[int64]UserRecord{}
	var userIdsToFetch []int64

	for _, userId := range userIds {
		cachedItem, hasCachedItem := w.cache.Get(fmt.Sprint(userId))

		if hasCachedItem {
			cachedItems[userId] = cachedItem.(UserRecord)
		} else {
			userIdsToFetch = append(userIdsToFetch, userId)
		}
	}

	finalResponse := GetUsersResponseChan{}

	if len(userIdsToFetch) == 0 {
		finalResponse.Items = cachedItems
		respCh <- finalResponse
		close(respCh)
		return respCh
	}

	respChan := w.baseWrapper.SendRpcRequest(w.serviceApiUrl, "GetUsersInternal", GetUsersRequest{
		UserIds: userIdsToFetch,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	go func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := GetUsersResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			data := map[int64]UserRecord{}

			if err := json.Unmarshal(resp.Result, &data); err != nil {
				result.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    w.baseWrapper.GetHostName(),
					ServiceName: w.serviceName,
				}
			} else {
				for userId, item := range data {
					w.cache.Set(fmt.Sprint(userId), item, cache.DefaultExpiration)
					cachedItems[userId] = item
				}
				result.Items = cachedItems
			}
		}

		respCh <- result
	}()

	return respCh
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

func (w *UserGoWrapper) AuthGuest(deviceId string, apmTransaction *apm.Transaction, forceLog bool) chan AuthGuestResponseChan {
	resChan := make(chan AuthGuestResponseChan, 2)

	go func() {
		link := fmt.Sprintf("%v/auth/guest", w.publicApiUrl)

		rpcInternalResponse := <-w.baseWrapper.SendRequestWithRpcResponseFromAnyService(link,
			"POST",
			"application/json",
			"auth guest",
			AuthGuestRequest{DeviceId: deviceId}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

		finalResponse := AuthGuestResponseChan{
			Error: rpcInternalResponse.Error,
		}
		if len(rpcInternalResponse.Result) > 0 {
			if err := json.Unmarshal(rpcInternalResponse.Result, &finalResponse); err != nil {
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
	}()

	return resChan
}

func (w *UserGoWrapper) GetBlockList(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetBlockListResponseChan {
	resChan := make(chan GetBlockListResponseChan, 2)

	go func() {
		link := fmt.Sprintf("%v/mobile/v1/user/block_list", w.publicApiUrl)

		rpcInternalResponse := <-w.baseWrapper.SendRequestWithRpcResponseFromAnyService(link,
			"POST",
			"application/json",
			"block list",
			GetBlockListRequest{UserIds: userIds}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

		finalResponse := GetBlockListResponseChan{
			Error: rpcInternalResponse.Error,
		}
		if len(rpcInternalResponse.Result) > 0 {
			if err := json.Unmarshal(rpcInternalResponse.Result, &finalResponse); err != nil {
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
	}()

	return resChan
}

func (w *UserGoWrapper) GetUserBlock(blockedTo int64, blockedBy int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUserBlockResponseChan {
	resChan := make(chan GetUserBlockResponseChan, 2)

	go func() {
		link := fmt.Sprintf("%v/mobile/v1/user/block_relations", w.publicApiUrl)

		rpcInternalResponse := <-w.baseWrapper.SendRequestWithRpcResponseFromAnyService(link,
			"POST",
			"application/json",
			"block relations",
			GetUserBlockRequest{
				BlockBy:   blockedBy,
				BlockedTo: blockedTo,
			}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

		finalResponse := GetUserBlockResponseChan{
			Error: rpcInternalResponse.Error,
		}
		if len(rpcInternalResponse.Result) > 0 {
			if err := json.Unmarshal(rpcInternalResponse.Result, &finalResponse); err != nil {
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
	}()

	return resChan
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
