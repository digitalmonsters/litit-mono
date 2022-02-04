package user

import (
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/digitalmonsters/go-common/wrappers"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"go.elastic.co/apm"
	"time"
)

type IUserWrapper interface {
	GetUsers(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersResponseChan
	GetUsersDetails(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersDetailsResponseChan
	GetProfileBulk(currentUserId int64, userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetProfileBulkResponseChan
}

//goland:noinspection GoNameStartsWithPackageName
type UserWrapper struct {
	baseWrapper    *wrappers.BaseWrapper
	defaultTimeout time.Duration
	apiUrl         string
	serviceName    string
	cache          *cache.Cache
}

func NewUserWrapper(config boilerplate.WrapperConfig) IUserWrapper {
	timeout := 5 * time.Second

	if config.TimeoutSec > 0 {
		timeout = time.Duration(config.TimeoutSec) * time.Second
	}

	if len(config.ApiUrl) == 0 {
		config.ApiUrl = "http://user-info"

		log.Warn().Msgf("Api Url is missing for User. Setting as default : %v", config.ApiUrl)
	}

	return &UserWrapper{
		baseWrapper:    wrappers.GetBaseWrapper(),
		defaultTimeout: timeout,
		apiUrl:         common.StripSlashFromUrl(config.ApiUrl),
		serviceName:    "user-info",
		cache:          cache.New(4*time.Minute, 5*time.Minute),
	}
}

func (w *UserWrapper) GetUsers(userIds []int64, apmTransaction *apm.Transaction,
	forceLog bool) chan GetUsersResponseChan {
	resChan := make(chan GetUsersResponseChan, 2)

	go func() {
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
			resChan <- finalResponse
			return
		}

		rpcInternalResponse := <-w.baseWrapper.SendRequestWithRpcResponseFromNodeJsService(fmt.Sprintf("%v/mobile/v1/profile", w.apiUrl),
			"POST",
			"application/json",
			"get users",
			GetUsersRequest{
				UserIds: userIdsToFetch,
			}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

		finalResponse.Error = rpcInternalResponse.Error

		if finalResponse.Error == nil && len(rpcInternalResponse.Result) > 0 {
			items := map[int64]UserRecord{}

			if err := json.Unmarshal(rpcInternalResponse.Result, &items); err != nil {
				finalResponse.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    w.baseWrapper.GetHostName(),
					ServiceName: w.serviceName,
				}
			} else {
				for userId, item := range items {
					w.cache.Set(fmt.Sprint(userId), item, cache.DefaultExpiration)
					cachedItems[userId] = item
				}

				finalResponse.Items = cachedItems
			}
		}

		resChan <- finalResponse
	}()

	return resChan
}

func (w *UserWrapper) GetUsersDetails(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersDetailsResponseChan {
	responseChan := make(chan GetUsersDetailsResponseChan, 2)
	response := GetUsersDetailsResponseChan{
		Items: make(map[int64]UserDetailRecord),
	}

	batchChannels := make([]chan UsersInternalChan, 0)

	for _, userId := range userIds {
		uid := userId
		resChan := make(chan UsersInternalChan, 2)
		batchChannels = append(batchChannels, resChan)

		go func() {
			finalResponse := UsersInternalChan{}

			rpcInternalResponse := <-w.baseWrapper.SendRequestWithRpcResponseFromNodeJsService(fmt.Sprintf("%v/mobile/v1/profile/%v/getProfile", w.apiUrl, uid),
				"GET",
				"application/json",
				"get users details",
				nil, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

			finalResponse.Error = rpcInternalResponse.Error

			if finalResponse.Error == nil && len(rpcInternalResponse.Result) > 0 {
				item := UserDetailRecord{}

				if err := json.Unmarshal(rpcInternalResponse.Result, &item); err != nil {
					finalResponse.Error = &rpc.RpcError{
						Code:        error_codes.GenericMappingError,
						Message:     err.Error(),
						Data:        nil,
						Hostname:    w.baseWrapper.GetHostName(),
						ServiceName: w.serviceName,
					}
				} else {
					finalResponse.UserDetailRecord = item
				}
			}

			resChan <- finalResponse
		}()
	}

	for _, c := range batchChannels {
		if internalResp := <-c; internalResp.Error != nil {
			apm_helper.CaptureApmError(errors.New(fmt.Sprintf("external service replied with error code [%v] snd message [%v]", internalResp.Error.Code, internalResp.Error.Message)), apmTransaction)
			continue
		} else {
			response.Items[internalResp.Id] = internalResp.UserDetailRecord
		}
	}

	responseChan <- response

	return responseChan
}

func (w *UserWrapper) GetProfileBulk(currentUserId int64, userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetProfileBulkResponseChan {
	resChan := make(chan GetProfileBulkResponseChan, 2)

	go func() {
		rpcInternalResponse := <-w.baseWrapper.SendRequestWithRpcResponseFromNodeJsService(fmt.Sprintf("%v/mobile/v1/profile/getProfiles", w.apiUrl),
			"POST",
			"application/json",
			"get profiles bulk",
			GetProfileBulkRequest{
				UserIds:       userIds,
				CurrentUserId: currentUserId,
			}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

		finalResponse := GetProfileBulkResponseChan{
			Error: rpcInternalResponse.Error,
		}

		if finalResponse.Error == nil && len(rpcInternalResponse.Result) > 0 {
			resp := map[int64]UserDetailRecord{}

			if err := json.Unmarshal(rpcInternalResponse.Result, &resp); err != nil {
				finalResponse.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    w.baseWrapper.GetHostName(),
					ServiceName: w.serviceName,
				}
			} else {
				finalResponse.Items = resp
			}
		}

		resChan <- finalResponse
	}()

	return resChan
}
