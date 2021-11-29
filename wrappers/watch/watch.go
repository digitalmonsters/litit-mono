package watch

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/digitalmonsters/go-common/wrappers"
	"go.elastic.co/apm"
	"time"
)

type IWatchWrapper interface {
	GetLastWatchesByUsers(userIds []int64, limitPerUser int, apmTransaction *apm.Transaction, forceLog bool) chan LastWatcherByUserResponseChan
}

//goland:noinspection GoNameStartsWithPackageName
type WatchWrapper struct {
	apiUrl         string
	baseWrapper    *wrappers.BaseWrapper
	defaultTimeout time.Duration
	serviceName    string
}

func NewWatchWrapper(apiUrl string) IWatchWrapper {
	return &WatchWrapper{baseWrapper: wrappers.GetBaseWrapper(), defaultTimeout: 5 * time.Second,
		apiUrl:      common.StripSlashFromUrl(apiUrl),
		serviceName: "watch-backend"}
}

type LastWatchesByUserRecord struct {
	ContentId  int64   `json:"content_id"`
	Duration   int     `json:"duration"`
	IsFullView bool    `json:"is_full_view"`
	Percent    float64 `json:"percent"`
}

type LastWatcherByUserResponseChan struct {
	Error *rpc.RpcError                       `json:"error"`
	Items map[int64][]LastWatchesByUserRecord `json:"items"`
}

type GetLatestWatchesByUserRequest struct {
	LimitPerUser int     `json:"limit_per_user"`
	UserIds      []int64 `json:"user_ids"`
	MinPercent   float64 `json:"min_percent"`
}

type CategoryInfo struct {
	Id         int64  `json:"id"`
	Name       string `json:"name"`
	ViewsCount int64  `json:"views_count"`
}

type GetCategoriesByViewsRequest struct {
	Limit  int64
	Offset int64
}

type GetCategoriesResponseChan struct {
	Error *rpc.RpcError  `json:"error"`
	Items []CategoryInfo `json:"items"`
}

func (w *WatchWrapper) GetLastWatchesByUsers(userIds []int64, limitPerUser int, apmTransaction *apm.Transaction,
	forceLog bool) chan LastWatcherByUserResponseChan {
	respCh := make(chan LastWatcherByUserResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "GetLastWatchesByUsers", GetLatestWatchesByUserRequest{
		LimitPerUser: limitPerUser,
		UserIds:      userIds,
	}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	w.baseWrapper.GetPool().Submit(func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := LastWatcherByUserResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			items := map[int64][]LastWatchesByUserRecord{}

			if err := json.Unmarshal(resp.Result, &items); err != nil {
				result.Error = &rpc.RpcError{
					Code:    error_codes.GenericMappingError,
					Message: err.Error(),
					Data:    nil,
				}
			} else {
				result.Items = items
			}
		}

		respCh <- result
	})

	return respCh
}

func (w *WatchWrapper) GetCategoriesByViews(limit int64, offset int64, apmTransaction *apm.Transaction) chan GetCategoriesResponseChan {
	respCh := make(chan GetCategoriesResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "GetCategoriesByViews", GetCategoriesByViewsRequest{
		Limit:  limit,
		Offset: offset,
	}, w.defaultTimeout, apmTransaction, w.serviceName, false)

	w.baseWrapper.GetPool().Submit(func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := GetCategoriesResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			items := make([]CategoryInfo, 0)

			if err := json.Unmarshal(resp.Result, &items); err != nil {
				result.Error = &rpc.RpcError{
					Code:    error_codes.GenericMappingError,
					Message: err.Error(),
					Data:    nil,
				}
			} else {
				result.Items = items
			}
		}

		respCh <- result
	})

	return respCh
}
