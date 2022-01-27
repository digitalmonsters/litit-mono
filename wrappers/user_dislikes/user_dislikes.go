package user_dislikes

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

type IUserDislikesWrapper interface {
	GetAllUserDislikes(userId int64, apmTransaction *apm.Transaction, forceLog bool) chan GetAllUserDislikesResponseChan
}

//goland:noinspection GoNameStartsWithPackageName
type UserDislikesWrapper struct {
	baseWrapper    *wrappers.BaseWrapper
	defaultTimeout time.Duration
	apiUrl         string
	serviceName    string
}

func NewUserDislikesWrapper(config boilerplate.WrapperConfig) IUserDislikesWrapper {
	timeout := 5 * time.Second

	if config.TimeoutSec > 0 {
		timeout = time.Duration(config.TimeoutSec) * time.Second
	}

	return &UserDislikesWrapper{
		baseWrapper:    wrappers.GetBaseWrapper(),
		defaultTimeout: timeout,
		apiUrl:         fmt.Sprintf("%v/rpc-service", common.StripSlashFromUrl(config.ApiUrl)),
		serviceName:    "content",
	}
}

func (w *UserDislikesWrapper) GetAllUserDislikes(userId int64, apmTransaction *apm.Transaction, forceLog bool) chan GetAllUserDislikesResponseChan {
	respCh := make(chan GetAllUserDislikesResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "InternalGetAllUserDislikes", GetAllUserDislikesRequest{
		UserId: userId,
	}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	go func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := GetAllUserDislikesResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			data := make([]int64, 0)

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
