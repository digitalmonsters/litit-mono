package like

import (
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/digitalmonsters/go-common/wrappers"
	"go.elastic.co/apm"
	"time"
)

type ILikeWrapper interface {
	GetLastLikesByUsers(userIds []int64, limitPerUser int, apmTransaction *apm.Transaction, forceLog bool) chan LastLikedByUserResponseChan
}

//goland:noinspection GoNameStartsWithPackageName
type LikeWrapper struct {
	baseWrapper    *wrappers.BaseWrapper
	defaultTimeout time.Duration
	apiUrl         string
	serviceName    string
}

func NewLikeWrapper(config boilerplate.WrapperConfig) ILikeWrapper {
	timeout := 5 * time.Second

	if config.TimeoutSec > 0 {
		timeout = time.Duration(config.TimeoutSec) * time.Second
	}

	return &LikeWrapper{
		baseWrapper:    wrappers.GetBaseWrapper(),
		defaultTimeout: timeout,
		apiUrl:         fmt.Sprintf("%v/rpc", common.StripSlashFromUrl(config.ApiUrl)),
		serviceName:    "likes",
	}
}

func (w *LikeWrapper) GetLastLikesByUsers(userIds []int64, limitPerUser int, apmTransaction *apm.Transaction,
	forceLog bool) chan LastLikedByUserResponseChan {
	respCh := make(chan LastLikedByUserResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "GetLastLikesByUsers", GetLatestLikedByUserRequest{
		LimitPerUser: limitPerUser,
		UserIds:      userIds,
	}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	w.baseWrapper.GetPool().Submit(func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := LastLikedByUserResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			items := map[int64][]LikeRecord{}

			if err := json.Unmarshal(resp.Result, &items); err != nil {
				result.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    w.baseWrapper.GetHostName(),
					ServiceName: w.serviceName,
				}
			} else {
				result.Items = items
			}
		}

		respCh <- result
	})

	return respCh
}

func (w *LikeWrapper) GetLikeContentUserByContentIdsInternal(contentIds []int64, userId int64, apmTransaction *apm.Transaction, forceLog bool) chan LikeContentUserByContentIdsResponseChan {
	respCh := make(chan LikeContentUserByContentIdsResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "GetLikeContentUserByContentIdsInternal", LikeContentUserByContentIdsRequest{
		UserId:     userId,
		ContentIds: contentIds,
	}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	w.baseWrapper.GetPool().Submit(func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := LikeContentUserByContentIdsResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			data := map[int64]bool{}

			if err := json.Unmarshal(resp.Result, &data); err != nil {
				result.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    w.baseWrapper.GetHostName(),
					ServiceName: w.serviceName,
				}
			} else {
				result.Data = data
			}
		}

		respCh <- result
	})

	return respCh
}
