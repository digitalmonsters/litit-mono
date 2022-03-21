package user_go

import (
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
	GetUsersDetails(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersDetailsResponseChan
	GetProfileBulk(currentUserId int64, userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetProfileBulkResponseChan
	GetUsersActiveThresholds(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersActiveThresholdsResponseChan
	GetUserIdsFilterByUsername(userIds []int64, searchQuery string, apmTransaction *apm.Transaction, forceLog bool) chan GetUserIdsFilterByUsernameResponseChan
	GetUsersTags(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersTagsResponseChan
	AuthGuest(deviceId string, apmTransaction *apm.Transaction, forceLog bool) chan AuthGuestResponseChan
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

func (u UserGoWrapper) GetUsers(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersResponseChan {
	respCh := make(chan GetUsersResponseChan, 2)

	cachedItems := map[int64]UserRecord{}
	var userIdsToFetch []int64

	for _, userId := range userIds {
		cachedItem, hasCachedItem := u.cache.Get(fmt.Sprint(userId))

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

	respChan := u.baseWrapper.SendRpcRequest(u.serviceApiUrl, "GetUsersInternal", GetUsersRequest{
		UserIds: userIdsToFetch,
	}, map[string]string{}, u.defaultTimeout, apmTransaction, u.serviceName, forceLog)

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
					Hostname:    u.baseWrapper.GetHostName(),
					ServiceName: u.serviceName,
				}
			} else {
				for userId, item := range data {
					u.cache.Set(fmt.Sprint(userId), item, cache.DefaultExpiration)
					cachedItems[userId] = item
				}
				result.Items = cachedItems
			}
		}

		respCh <- result
	}()

	return respCh
}

func (u UserGoWrapper) GetUsersDetails(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersDetailsResponseChan {
	respCh := make(chan GetUsersDetailsResponseChan, 2)

	respChan := u.baseWrapper.SendRpcRequest(u.serviceApiUrl, "GetUsersDetailsInternal", GetUsersDetailRequest{
		UserIds: userIds,
	}, map[string]string{}, u.defaultTimeout, apmTransaction, u.serviceName, forceLog)

	go func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := GetUsersDetailsResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			data := map[int64]UserDetailRecord{}

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

func (u UserGoWrapper) GetProfileBulk(currentUserId int64, userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetProfileBulkResponseChan {
	respCh := make(chan GetProfileBulkResponseChan, 2)

	respChan := u.baseWrapper.SendRpcRequest(u.serviceApiUrl, "GetProfileBulkInternal", GetProfileBulkRequest{
		CurrentUserId: currentUserId,
		UserIds:       userIds,
	}, map[string]string{}, u.defaultTimeout, apmTransaction, u.serviceName, forceLog)

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

func (u UserGoWrapper) GetUsersActiveThresholds(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersActiveThresholdsResponseChan {
	respCh := make(chan GetUsersActiveThresholdsResponseChan, 2)

	respChan := u.baseWrapper.SendRpcRequest(u.serviceApiUrl, "GetUsersActiveThresholds", GetUsersActiveThresholdsRequest{
		UserIds: userIds,
	}, map[string]string{}, u.defaultTimeout, apmTransaction, u.serviceName, forceLog)

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

func (u UserGoWrapper) GetUserIdsFilterByUsername(userIds []int64, searchQuery string, apmTransaction *apm.Transaction, forceLog bool) chan GetUserIdsFilterByUsernameResponseChan {
	respCh := make(chan GetUserIdsFilterByUsernameResponseChan, 2)

	respChan := u.baseWrapper.SendRpcRequest(u.serviceApiUrl, "GetUserIdsFilterByUsername", GetUserIdsFilterByUsernameRequest{
		UserIds:     userIds,
		SearchQuery: searchQuery,
	}, map[string]string{}, u.defaultTimeout, apmTransaction, u.serviceName, forceLog)

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
					Hostname:    u.baseWrapper.GetHostName(),
					ServiceName: u.serviceName,
				}
			} else {
				result.UserIds = data
			}
		}

		respCh <- result
	}()

	return respCh
}

func (u UserGoWrapper) GetUsersTags(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersTagsResponseChan {
	respCh := make(chan GetUsersTagsResponseChan, 2)

	respChan := u.baseWrapper.SendRpcRequest(u.serviceApiUrl, "GetUsersTags", GetUsersTagsRequest{
		UserIds: userIds,
	}, map[string]string{}, u.defaultTimeout, apmTransaction, u.serviceName, forceLog)

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

func (u *UserGoWrapper) AuthGuest(deviceId string, apmTransaction *apm.Transaction, forceLog bool) chan AuthGuestResponseChan {
	resChan := make(chan AuthGuestResponseChan, 2)

	go func() {
		link := fmt.Sprintf("%v/auth/guest", u.publicApiUrl)

		rpcInternalResponse := <-u.baseWrapper.SendRequestWithRpcResponseFromAnyService(link,
			"POST",
			"application/json",
			"auth guest",
			AuthGuestRequest{DeviceId: deviceId}, map[string]string{}, u.defaultTimeout, apmTransaction, u.serviceName, forceLog)

		finalResponse := AuthGuestResponseChan{
			Error: rpcInternalResponse.Error,
		}
		if len(rpcInternalResponse.Result) > 0 {
			if err := json.Unmarshal(rpcInternalResponse.Result, &finalResponse); err != nil {
				finalResponse.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    u.baseWrapper.GetHostName(),
					ServiceName: u.serviceName,
				}
			}
		}

		resChan <- finalResponse
	}()

	return resChan
}
