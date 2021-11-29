package hashtag

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

type IHashtagWrapper interface {
	GetHashtagsInternal(hashtags []string, limit int, offset int, apmTransaction *apm.Transaction, forceLog bool) chan HashtagsGetInternalResponseChan
}

func NewHashtagWrapper(apiUrl string) IHashtagWrapper {
	return &Wrapper{
		baseWrapper:    wrappers.GetBaseWrapper(),
		defaultTimeout: 5 * time.Second,
		apiUrl:         common.StripSlashFromUrl(apiUrl),
		serviceName:    "content-backend"}
}

func (w *Wrapper) GetHashtagsInternal(hashtags []string, limit int, offset int, apmTransaction *apm.Transaction, forceLog bool) chan HashtagsGetInternalResponseChan {
	respCh := make(chan HashtagsGetInternalResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "GetHashtagsInternal", GetHashtagsInternalRequest{
		Hashtags: hashtags,
		Limit:    limit,
		Offset:   offset,
	}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	w.baseWrapper.GetPool().Submit(func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := HashtagsGetInternalResponseChan{
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
