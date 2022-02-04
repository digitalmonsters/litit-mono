package user_likes

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

type IUserLikesWrapper interface {
	GetUserLikes(userId int64, limit int, offset int, apmTransaction *apm.Transaction, forceLog bool) chan GetUserLikesResponseChan
}

//goland:noinspection GoNameStartsWithPackageName
type UserLikesWrapper struct {
	baseWrapper    *wrappers.BaseWrapper
	defaultTimeout time.Duration
	apiUrl         string
	serviceName    string
}

func NewUserLikesWrapper(config boilerplate.WrapperConfig) IUserLikesWrapper {
	timeout := 5 * time.Second

	if config.TimeoutSec > 0 {
		timeout = time.Duration(config.TimeoutSec) * time.Second
	}

	if len(config.ApiUrl) == 0 {
		config.ApiUrl = "http://content"

		log.Warn().Msgf("Api Url is missing for UserLikes. Setting as default : %v", config.ApiUrl)
	}

	return &UserLikesWrapper{
		baseWrapper:    wrappers.GetBaseWrapper(),
		defaultTimeout: timeout,
		apiUrl:         fmt.Sprintf("%v/rpc-service", common.StripSlashFromUrl(config.ApiUrl)),
		serviceName:    "content",
	}
}

func (w *UserLikesWrapper) GetUserLikes(userId int64, limit int, offset int, apmTransaction *apm.Transaction, forceLog bool) chan GetUserLikesResponseChan {
	respCh := make(chan GetUserLikesResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "InternalGetUserLikes", GetUserLikesRequest{
		UserId: userId,
		Limit:  limit,
		Offset: offset,
	}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	go func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := GetUserLikesResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			data := LikedContent{}

			if err := json.Unmarshal(resp.Result, &data); err != nil {
				result.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    w.baseWrapper.GetHostName(),
					ServiceName: w.serviceName,
				}
			} else {
				result.Data = &data
			}
		}

		respCh <- result
	}()

	return respCh
}
