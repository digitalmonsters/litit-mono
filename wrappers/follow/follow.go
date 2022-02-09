package follow

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

type IFollowWrapper interface {
	GetUserFollowingRelationBulk(userId int64, requestUserIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUserFollowingRelationBulkResponseChan
	GetUserFollowingRelation(userId int64, requestUserId int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUserFollowingRelationResponseChan
	GetUserFollowers(userId int64, pageState string, limit int, apmTransaction *apm.Transaction, forceLog bool) chan GetUserFollowersResponseChan
	GetFollowersCount(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetFollowersCountResponseChan
}

//goland:noinspection GoNameStartsWithPackageName
type FollowWrapper struct {
	baseWrapper    *wrappers.BaseWrapper
	defaultTimeout time.Duration
	apiUrl         string
	serviceName    string
}

func NewFollowWrapper(config boilerplate.WrapperConfig) IFollowWrapper {
	timeout := 5 * time.Second

	if config.TimeoutSec > 0 {
		timeout = time.Duration(config.TimeoutSec) * time.Second
	}

	if len(config.ApiUrl) == 0 {
		config.ApiUrl = "http://follows"

		log.Warn().Msgf("Api Url is missing for Follow. Setting as default : %v", config.ApiUrl)
	}

	return &FollowWrapper{
		baseWrapper:    wrappers.GetBaseWrapper(),
		defaultTimeout: timeout,
		apiUrl:         fmt.Sprintf("%v/rpc-service", common.StripSlashFromUrl(config.ApiUrl)),
		serviceName:    "follows",
	}
}

func (w *FollowWrapper) GetUserFollowingRelationBulk(userId int64, requestUserIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUserFollowingRelationBulkResponseChan {
	respCh := make(chan GetUserFollowingRelationBulkResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "InternalUserFollowRelationBulk", GetUserFollowingRelationBulkRequest{
		UserId:         userId,
		RequestUserIds: requestUserIds,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	go func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := GetUserFollowingRelationBulkResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			res := UserFollowingRelationResponse{
				Data: map[int64]RelationData{},
			}

			if err := json.Unmarshal(resp.Result, &res); err != nil {
				result.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    w.baseWrapper.GetHostName(),
					ServiceName: w.serviceName,
				}
			} else {
				result.Data = res.Data
			}
		}

		respCh <- result
	}()

	return respCh
}

func (w *FollowWrapper) GetUserFollowingRelation(userId int64, requestUserId int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUserFollowingRelationResponseChan {
	respCh := make(chan GetUserFollowingRelationResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "InternalUserFollowRelation", GetUserFollowingRelationRequest{
		UserId:        userId,
		RequestUserId: requestUserId,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	go func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := GetUserFollowingRelationResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			data := RelationData{}

			if err := json.Unmarshal(resp.Result, &data); err != nil {
				result.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    w.baseWrapper.GetHostName(),
					ServiceName: w.serviceName,
				}
			} else {
				result.IsFollower = data.IsFollower
				result.IsFollowing = data.IsFollowing
			}
		}

		respCh <- result
	}()

	return respCh
}

func (w *FollowWrapper) GetUserFollowers(userId int64, pageState string, limit int, apmTransaction *apm.Transaction, forceLog bool) chan GetUserFollowersResponseChan {
	respCh := make(chan GetUserFollowersResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "InternalGetUserFollowers", GetUserFollowersRequest{
		UserId:    userId,
		PageState: pageState,
		Limit:     limit,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	go func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := GetUserFollowersResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			data := GetUserFollowersResponse{}

			if err := json.Unmarshal(resp.Result, &data); err != nil {
				result.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    w.baseWrapper.GetHostName(),
					ServiceName: w.serviceName,
				}
			} else {
				result.FollowerIds = data.FollowerIds
				result.PageState = data.PageState
			}
		}

		respCh <- result
	}()

	return respCh
}

func (w *FollowWrapper) GetFollowersCount(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetFollowersCountResponseChan {
	respCh := make(chan GetFollowersCountResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "InternalGetFollowersCount", GetFollowersCountRequest{
		UserIds: userIds,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	go func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := GetFollowersCountResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			data := make(map[int64]int64)

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
