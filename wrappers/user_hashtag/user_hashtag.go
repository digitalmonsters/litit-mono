package user_hashtag

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

type IUserHashtagWrapper interface {
	GetUserHashtagSubscriptionStateBulk(hashtags []string, userId int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUserHashtagSubscriptionStateResponseChan
}

//goland:noinspection GoNameStartsWithPackageName
type UserHashtagWrapper struct {
	baseWrapper    *wrappers.BaseWrapper
	defaultTimeout time.Duration
	apiUrl         string
	serviceName    string
}

func NewUserCategoryWrapper(config boilerplate.WrapperConfig) IUserHashtagWrapper {
	timeout := 5 * time.Second

	if config.TimeoutSec > 0 {
		timeout = time.Duration(config.TimeoutSec) * time.Second
	}

	return &UserHashtagWrapper{
		baseWrapper:    wrappers.GetBaseWrapper(),
		defaultTimeout: timeout,
		apiUrl:         fmt.Sprintf("%v/rpc", common.StripSlashFromUrl(config.ApiUrl)),
		serviceName:    "user-hashtags",
	}
}

func (w *UserHashtagWrapper) GetUserHashtagSubscriptionStateBulk(hashtags []string, userId int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUserHashtagSubscriptionStateResponseChan {
	respCh := make(chan GetUserHashtagSubscriptionStateResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "GetInternalUserHashtagSubscriptionStateBulk", GetUserHashtagSubscriptionStateBulkRequest{
		UserId:   userId,
		Hashtags: hashtags,
	}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	go func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := GetUserHashtagSubscriptionStateResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			data := map[string]bool{}

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
