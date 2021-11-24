package content

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/digitalmonsters/go-common/wrappers"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
	"time"
)

type SimpleContent struct {
	Id            int64    `json:"id"`
	Duration      int      `json:"duration"`
	AgeRestricted bool     `json:"age_restricted"`
	AuthorId      int64    `json:"author_id"`
	CategoryId    null.Int `json:"category_id"`
	Hashtags      []string `json:"hashtags"`
}

//goland:noinspection ALL
type ContentGetInternalResponseChan struct {
	Error *rpc.RpcError           `json:"error"`
	Items map[int64]SimpleContent `json:"items"`
}

type ContentGetInternalRequest struct {
	IncludeDeleted bool    `json:"include_deleted"`
	ContentIds     []int64 `json:"content_ids"`
}

type IContentWrapper interface {
	GetInternal(contentIds []int64, includeDeleted bool, apmTransaction *apm.Transaction, forceLog bool) chan ContentGetInternalResponseChan
}

//goland:noinspection GoNameStartsWithPackageName
type ContentWrapper struct {
	baseWrapper    *wrappers.BaseWrapper
	defaultTimeout time.Duration
	apiUrl         string
	serviceName    string
}

func NewContentWrapper(apiUrl string) IContentWrapper {
	return &ContentWrapper{baseWrapper: wrappers.GetBaseWrapper(), defaultTimeout: 5 * time.Second,
		apiUrl: common.StripSlashFromUrl(apiUrl),
		serviceName: "content-backend"}
}

func (w *ContentWrapper) GetInternal(contentIds []int64, includeDeleted bool, apmTransaction *apm.Transaction, forceLog bool) chan ContentGetInternalResponseChan {
	respCh := make(chan ContentGetInternalResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "ContentGetInternal", ContentGetInternalRequest{
		ContentIds:     contentIds,
		IncludeDeleted: includeDeleted,
	}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	w.baseWrapper.GetPool().Submit(func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := ContentGetInternalResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			items := map[int64]SimpleContent{}

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
