package hashtag

import (
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/digitalmonsters/go-common/wrappers"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
	"time"
)

type Wrapper struct {
	baseWrapper    *wrappers.BaseWrapper
	defaultTimeout time.Duration
	apiUrl         string
	serviceName    string
}

type IHashtagWrapper interface {
	GetHashtagsInternal(hashtags []string, omitHashtags []string, limit int, offset int, withViews null.Bool, apmTransaction *apm.Transaction,
		shouldHaveValidContent bool, forceLog bool) chan HashtagsGetInternalResponseChan
}

func NewHashtagWrapper(config boilerplate.WrapperConfig) IHashtagWrapper {
	timeout := 5 * time.Second

	if config.TimeoutSec > 0 {
		timeout = time.Duration(config.TimeoutSec) * time.Second
	}

	return &Wrapper{
		baseWrapper:    wrappers.GetBaseWrapper(),
		defaultTimeout: timeout,
		apiUrl:         fmt.Sprintf("%v/rpc-service", common.StripSlashFromUrl(config.ApiUrl)),
		serviceName:    "content",
	}
}

func (w *Wrapper) GetHashtagsInternal(hashtags []string, omitHashtags []string, limit int, offset int, withViews null.Bool,
	apmTransaction *apm.Transaction, shouldHaveValidContent bool, forceLog bool) chan HashtagsGetInternalResponseChan {
	respCh := make(chan HashtagsGetInternalResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "GetHashtagsInternal", GetHashtagsInternalRequest{
		Hashtags:               hashtags,
		OmitHashtags:           omitHashtags,
		Limit:                  limit,
		WithViews:              withViews,
		Offset:                 offset,
		ShouldHaveValidContent: shouldHaveValidContent,
	}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	go func() {
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
	}()

	return respCh
}
