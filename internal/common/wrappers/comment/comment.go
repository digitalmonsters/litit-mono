package comment

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

type ICommentWrapper interface {
	GetCommentsInfoById(commentIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetCommentsInfoByIdResponseChan
}

type CommentWrapper struct {
	defaultTimeout time.Duration
	apiUrl         string
	serviceName    string
	baseWrapper    *wrappers.BaseWrapper
}

func NewCommentWrapper(config boilerplate.WrapperConfig) ICommentWrapper {
	timeout := 5 * time.Second

	if config.TimeoutSec > 0 {
		timeout = time.Duration(config.TimeoutSec) * time.Second
	}

	if len(config.ApiUrl) == 0 {
		config.ApiUrl = "http://comments"

		log.Warn().Msgf("Api Url is missing for Comment. Setting as default : %v", config.ApiUrl)
	}

	return &CommentWrapper{
		defaultTimeout: timeout,
		apiUrl:         fmt.Sprintf("%v/rpc-service", common.StripSlashFromUrl(config.ApiUrl)),
		serviceName:    "comments",
		baseWrapper:    wrappers.GetBaseWrapper(),
	}
}

func (u CommentWrapper) GetCommentsInfoById(commentIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetCommentsInfoByIdResponseChan {
	respCh := make(chan GetCommentsInfoByIdResponseChan, 2)

	respChan := u.baseWrapper.SendRpcRequest(u.apiUrl, "GetCommentsInfoById", GetCommentsInfoByIdRequest{
		CommentIds: commentIds,
	}, map[string]string{}, u.defaultTimeout, apmTransaction, u.serviceName, forceLog)

	go func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := GetCommentsInfoByIdResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			var data = make(map[int64]CommentsInfoById)

			if err := json.Unmarshal(resp.Result, &data); err != nil {
				result.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    u.baseWrapper.GetHostName(),
					ServiceName: u.serviceName,
				}
			} else {
				result.Items = data
			}
		}

		respCh <- result
	}()

	return respCh
}
