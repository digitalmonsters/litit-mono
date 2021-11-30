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

type SimpleHashtag struct {
	Name       string `json:"name"`
	ViewsCount int    `json:"views_count"`
}

type ResponseData struct {
	Items      []SimpleHashtag `json:"items"`
	TotalCount int64           `json:"total_count"`
}

type HashtagsGetInternalResponseChan struct {
	Error *rpc.RpcError `json:"error"`
	Data  *ResponseData `json:"data"`
}

type GetHashtagsInternalRequest struct {
	Hashtags []string `json:"hashtags"`
	Limit    int      `json:"limit"`
	Offset   int      `json:"offset"`
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

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "GetCategoryInternal", GetHashtagsInternalRequest{
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
					Hostname: w.baseWrapper.GetHostName(),
				}
			} else {
				result.Data = data
			}
		}

		respCh <- result
	})

	return respCh
}
