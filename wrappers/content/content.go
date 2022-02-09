package content

import (
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/digitalmonsters/go-common/wrappers"
	"github.com/rs/zerolog/log"
	"go.elastic.co/apm"
	"time"
)

type IContentWrapper interface {
	GetInternal(contentIds []int64, includeDeleted bool, apmTransaction *apm.Transaction, forceLog bool) chan ContentGetInternalResponseChan
	GetTopNotFollowingUsers(userId int64, limit int, apmTransaction *apm.Transaction, forceLog bool) chan GetTopNotFollowingUsersResponseChan
}

//goland:noinspection GoNameStartsWithPackageName
type ContentWrapper struct {
	baseWrapper    *wrappers.BaseWrapper
	defaultTimeout time.Duration
	apiUrl         string
	serviceName    string
}

func NewContentWrapper(config boilerplate.WrapperConfig) IContentWrapper {
	timeout := 5 * time.Second

	if config.TimeoutSec > 0 {
		timeout = time.Duration(config.TimeoutSec) * time.Second
	}

	if len(config.ApiUrl) == 0 {
		config.ApiUrl = "http://content"

		log.Warn().Msgf("Api Url is missing for Content. Setting as default : %v", config.ApiUrl)
	}

	return &ContentWrapper{
		baseWrapper:    wrappers.GetBaseWrapper(),
		defaultTimeout: timeout,
		apiUrl:         fmt.Sprintf("%v/rpc-service", common.StripSlashFromUrl(config.ApiUrl)),
		serviceName:    "content",
	}
}

func (w *ContentWrapper) GetInternal(contentIds []int64, includeDeleted bool, apmTransaction *apm.Transaction, forceLog bool) chan ContentGetInternalResponseChan {
	respCh := make(chan ContentGetInternalResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "ContentGetInternal", ContentGetInternalRequest{
		ContentIds:     contentIds,
		IncludeDeleted: includeDeleted,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	go func() {
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
	}()

	return respCh
}

func (w *ContentWrapper) GetTopNotFollowingUsers(userId int64, limit int, apmTransaction *apm.Transaction, forceLog bool) chan GetTopNotFollowingUsersResponseChan {
	respCh := make(chan GetTopNotFollowingUsersResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "GetTopNotFollowingUsers", GetTopNotFollowingUsersRequest{
		UserId: userId,
		Limit:  limit,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	go func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := GetTopNotFollowingUsersResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			var items []int64

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
	}()

	return respCh
}
