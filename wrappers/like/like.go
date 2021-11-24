package like

import (
	"encoding/json"
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

func NewLikeWrapper(apiUrl string) ILikeWrapper {
	return &LikeWrapper{baseWrapper: wrappers.GetBaseWrapper(), defaultTimeout: 5 * time.Second, apiUrl: common.StripSlashFromUrl(apiUrl),
		serviceName: "like-backend"}
}

type LastLikedByUserResponseChan struct {
	Error *rpc.RpcError          `json:"error"`
	Items map[int64][]LikeRecord `json:"items"`
}

//goland:noinspection GoNameStartsWithPackageName
type LikeRecord struct {
	ContentId int64 `json:"content_id"`
}

type GetLatestLikedByUserRequest struct {
	LimitPerUser int     `json:"limit_per_user"`
	UserIds      []int64 `json:"user_ids"`
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
