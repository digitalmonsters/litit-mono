package user_block

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

type IUserBlockWrapper interface {
	GetBlockList(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetBlockListResponseChan
	GetUserBlock(blockedTo int64, blockedBy int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUserBlockResponseChan
}

//goland:noinspection GoNameStartsWithPackageName
type UserBlockWrapper struct {
	baseWrapper    *wrappers.BaseWrapper
	defaultTimeout time.Duration
	apiUrl         string
	serviceName    string
}

func NewUserBlockWrapper(config boilerplate.WrapperConfig) IUserBlockWrapper {
	timeout := 5 * time.Second

	if config.TimeoutSec > 0 {
		timeout = time.Duration(config.TimeoutSec) * time.Second
	}

	if len(config.ApiUrl) == 0 {
		config.ApiUrl = "http://user-block"

		log.Warn().Msgf("Api Url is missing for UserBlock. Setting as default : %v", config.ApiUrl)
	}

	return &UserBlockWrapper{
		baseWrapper:    wrappers.GetBaseWrapper(),
		defaultTimeout: timeout,
		apiUrl:         common.StripSlashFromUrl(config.ApiUrl),
		serviceName:    "user-block",
	}
}

func (w *UserBlockWrapper) GetBlockList(userIds []int64, apmTransaction *apm.Transaction,
	forceLog bool) chan GetBlockListResponseChan {
	resChan := make(chan GetBlockListResponseChan, 2)

	go func() {
		finalResponse := GetBlockListResponseChan{}

		rpcInternalResponse := <-w.baseWrapper.SendRequestWithRpcResponseFromNodeJsService(
			fmt.Sprintf("%v/mobile/v1/user/block_list", w.apiUrl),
			"POST",
			"application/json",
			"get block list",
			GetBlockListRequest{
				UserIds: userIds,
			}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

		finalResponse.Error = rpcInternalResponse.Error

		if finalResponse.Error == nil && len(rpcInternalResponse.Result) > 0 {
			var userBlockData map[int64][]int64

			if err := json.Unmarshal(rpcInternalResponse.Result, &userBlockData); err != nil {
				finalResponse.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    w.baseWrapper.GetHostName(),
					ServiceName: w.serviceName,
				}
			} else {
				finalResponse.Data = userBlockData
			}
		}

		resChan <- finalResponse
	}()

	return resChan
}

func (w *UserBlockWrapper) GetUserBlock(blockedTo int64, blockedBy int64, apmTransaction *apm.Transaction,
	forceLog bool) chan GetUserBlockResponseChan {
	resChan := make(chan GetUserBlockResponseChan, 2)

	go func() {
		finalResponse := GetUserBlockResponseChan{}

		rpcInternalResponse := <-w.baseWrapper.SendRequestWithRpcResponseFromNodeJsService(
			fmt.Sprintf("%v/mobile/v1/user/block_relations", w.apiUrl),
			"POST",
			"application/json",
			"get user block",
			GetUserBlockRequest{
				BlockBy:   blockedBy,
				BlockedTo: blockedTo,
			}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

		finalResponse.Error = rpcInternalResponse.Error

		if finalResponse.Error == nil && len(rpcInternalResponse.Result) > 0 {
			var userBlockData UserBlockData

			if err := json.Unmarshal(rpcInternalResponse.Result, &userBlockData); err != nil {
				finalResponse.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    w.baseWrapper.GetHostName(),
					ServiceName: w.serviceName,
				}
			} else {
				finalResponse.Data = userBlockData
			}
		}

		resChan <- finalResponse
	}()

	return resChan
}
