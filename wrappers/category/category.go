package category

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/digitalmonsters/go-common/wrappers"
	"go.elastic.co/apm"
	"time"
)

type Wrapper struct {
	baseWrapper    *wrappers.BaseWrapper
	defaultTimeout time.Duration
	apiUrl         string
	serviceName    string
}

type ICategoryWrapper interface {
	GetCategoryInternal(categoryIds []int64, limit int, offset int, excludeRoot bool, apmTransaction *apm.Transaction, forceLog bool) chan CategoryGetInternalResponseChan
}

func NewCategoryWrapper(apiUrl string) ICategoryWrapper {
	return &Wrapper{
		baseWrapper:    wrappers.GetBaseWrapper(),
		defaultTimeout: 5 * time.Second,
		apiUrl:         common.StripSlashFromUrl(apiUrl),
		serviceName:    "content-backend"}
}

func (w *Wrapper) GetCategoryInternal(categoryIds []int64, limit int, offset int, excludeRoot bool, apmTransaction *apm.Transaction, forceLog bool) chan CategoryGetInternalResponseChan {
	respCh := make(chan CategoryGetInternalResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "GetCategoryInternal", GetCategoryInternalRequest{
		CategoryIds: categoryIds,
		Limit:       limit,
		Offset:      offset,
		ExcludeRoot: excludeRoot,
	}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	w.baseWrapper.GetPool().Submit(func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := CategoryGetInternalResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			data := &ResponseData{}

			if err := json.Unmarshal(resp.Result, &data); err != nil {
				result.Error = &rpc.RpcError{
					Code:    error_codes.GenericMappingError,
					Message: err.Error(),
					Data:    nil,
				}
			} else {
				result.Data = data
			}
		}

		respCh <- result
	})

	return respCh
}
