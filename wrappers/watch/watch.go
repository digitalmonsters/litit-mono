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
