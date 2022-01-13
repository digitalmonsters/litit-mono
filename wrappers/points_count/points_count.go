package points_count

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

type IPointsCountWrapper interface {
	GetPointsCount(contentIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetPointsCountResponseChan
}

//goland:noinspection GoNameStartsWithPackageName
type PointsCountWrapper struct {
	baseWrapper    *wrappers.BaseWrapper
	defaultTimeout time.Duration
	apiUrl         string
	serviceName    string
	cache          *cache.Cache
}

func NewPointsCountWrapper(config boilerplate.WrapperConfig) IPointsCountWrapper {
	timeout := 5 * time.Second

	if config.TimeoutSec > 0 {
		timeout = time.Duration(config.TimeoutSec) * time.Second
	}

	return &PointsCountWrapper{
		baseWrapper:    wrappers.GetBaseWrapper(),
		defaultTimeout: timeout,
		apiUrl:         common.StripSlashFromUrl(config.ApiUrl),
		serviceName:    "rewards-and-vault",
		cache:          cache.New(4*time.Minute, 5*time.Minute),
	}
}

func (w *PointsCountWrapper) GetPointsCount(contentIds []int64, apmTransaction *apm.Transaction,
	forceLog bool) chan GetPointsCountResponseChan {
	resChan := make(chan GetPointsCountResponseChan, 2)

	go func() {
		cachedItems := map[int64]PointsCountRecord{}
		var contentIdsToFetch []int64

		for _, contentId := range contentIds {
			cachedItem, hasCachedItem := w.cache.Get(fmt.Sprint(contentId))

			if hasCachedItem {
				cachedItems[contentId] = cachedItem.(PointsCountRecord)
			} else {
				contentIdsToFetch = append(contentIdsToFetch, contentId)
			}
		}

		finalResponse := GetPointsCountResponseChan{}

		if len(contentIdsToFetch) == 0 {
			finalResponse.Items = cachedItems
			resChan <- finalResponse
			return
		}

		rpcInternalResponse := <-w.baseWrapper.SendRequestWithRpcResponseFromNodeJsService(fmt.Sprintf("%v/mobile/v1/content/points", w.apiUrl),
			"POST",
			"application/json",
			"get points count",
			GetPointsCountRequest{
				ContentIds: contentIdsToFetch,
			}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

		finalResponse.Error = rpcInternalResponse.Error

		if finalResponse.Error == nil && len(rpcInternalResponse.Result) > 0 {
			items := map[int64]PointsCountRecord{}

			if err := json.Unmarshal(rpcInternalResponse.Result, &items); err != nil {
				finalResponse.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    w.baseWrapper.GetHostName(),
					ServiceName: w.serviceName,
				}
			} else {
				for contentId, item := range items {
					w.cache.Set(fmt.Sprint(contentId), item, cache.DefaultExpiration)
					cachedItems[contentId] = item
				}

				finalResponse.Items = cachedItems
			}
		}

		resChan <- finalResponse
	}()

	return resChan
}
