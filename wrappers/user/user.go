package user

import (
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/digitalmonsters/go-common/wrappers"
	"github.com/patrickmn/go-cache"
	"go.elastic.co/apm"
	"time"
)

type IUserWrapper interface {
	GetUsers(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersResponseChan
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

	return &UserWrapper{
		baseWrapper:    wrappers.GetBaseWrapper(),
		defaultTimeout: timeout,
		apiUrl:         common.StripSlashFromUrl(config.ApiUrl),
		serviceName:    "user-backend",
		cache:          cache.New(4*time.Minute, 5*time.Minute),
	}
}

func (w *UserWrapper) GetUsers(userIds []int64, apmTransaction *apm.Transaction,
	forceLog bool) chan GetUsersResponseChan {
	resChan := make(chan GetUsersResponseChan, 2)

	w.baseWrapper.GetPool().Submit(func() {
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
					Code:     error_codes.GenericMappingError,
					Message:  err.Error(),
					Data:     nil,
					Hostname: w.baseWrapper.GetHostName(),
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
	})

	return resChan
}
