package go_tokenomics

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/filters"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/digitalmonsters/go-common/wrappers"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
	"go.elastic.co/apm"
)

type Wrapper struct {
	baseWrapper    *wrappers.BaseWrapper
	defaultTimeout time.Duration
	apiUrl         string
	serviceName    string
}

type IGoTokenomicsWrapper interface {
	GetUsersTokenomicsInfo(userIds []int64, filters []filters.Filter, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]UserTokenomicsInfo]
	GetWithdrawalsAmountsByAdminIds(adminIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]decimal.Decimal]
	GetContentEarningsTotalByContentIds(contentIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetContentEarningsTotalByContentIdsResponseChan
	GetTokenomicsStatsByUserId(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetTokenomicsStatsByUserIdResponseChan
	GetConfigProperties(properties []string, apmTransaction *apm.Transaction, forceLog bool) chan GetConfigPropertiesResponseChan
	GetReferralsInfo(referrerId int64, referralIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[GetReferralInfoResponse]
	GetActivitiesInfo(userId int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[GetActivitiesInfoResponse]
	CreateBotViews(botViews map[int64][]int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[any]
	WriteOffUserTokensForAd(userId int64, adCampaignId int64, amount decimal.Decimal, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[any]
	GetReferralsProgressInfo(referrerId int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[GetReferralProgressInfoResponse]
	GetMyReferredUsersWatchedVideoInfo(referrerId, page, count int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[GetMyReferredUsersWatchedVideoInfoResponse]
	DeductVaultPointsForIntroFeed(userId int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[DeductVaultPointsForIntroFeedResponse]
	AddPointsToVault(userId int64, points decimal.Decimal, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[AddPointsToVaultResponse]
}

func NewGoTokenomicsWrapper(config boilerplate.WrapperConfig) IGoTokenomicsWrapper {
	timeout := 5 * time.Second

	if config.TimeoutSec > 0 {
		timeout = time.Duration(config.TimeoutSec) * time.Second
	}

	if len(config.ApiUrl) == 0 {
		config.ApiUrl = "http://go-tokenomics"

		log.Warn().Msgf("Api Url is missing for GoTokenomics. Setting as default : %v", config.ApiUrl)
	}

	return &Wrapper{
		baseWrapper:    wrappers.GetBaseWrapper(),
		defaultTimeout: timeout,
		apiUrl:         fmt.Sprintf("%v/rpc-service", common.StripSlashFromUrl(config.ApiUrl)),
		serviceName:    "go tokenomics",
	}
}

func (w *Wrapper) GetUsersTokenomicsInfo(userIds []int64, filters []filters.Filter, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]UserTokenomicsInfo] {
	return wrappers.ExecuteRpcRequestAsync[map[int64]UserTokenomicsInfo](w.baseWrapper, w.apiUrl, "GetUsersTokenomicsInfo", GetUsersTokenomicsInfoRequest{
		UserIds: userIds,
		Filters: filters,
	}, map[string]string{}, w.defaultTimeout, apm.TransactionFromContext(ctx), w.serviceName, forceLog)
}

func (w *Wrapper) GetWithdrawalsAmountsByAdminIds(adminIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]decimal.Decimal] {
	return wrappers.ExecuteRpcRequestAsync[map[int64]decimal.Decimal](w.baseWrapper, w.apiUrl, "GetWithdrawalsAmountsByAdminIds", GetWithdrawalsAmountsByAdminIdsRequest{
		AdminIds: adminIds,
	}, map[string]string{}, w.defaultTimeout, apm.TransactionFromContext(ctx), w.serviceName, forceLog)
}

func (w *Wrapper) GetContentEarningsTotalByContentIds(contentIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetContentEarningsTotalByContentIdsResponseChan {
	respCh := make(chan GetContentEarningsTotalByContentIdsResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "GetContentEarningsTotal", GetContentEarningsTotalByContentIdsRequest{
		ContentIds: contentIds,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	go func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := GetContentEarningsTotalByContentIdsResponseChan{
			Error: resp.Error,
		}
		if len(resp.Result) > 0 {
			data := map[int64]decimal.Decimal{}

			if err := json.Unmarshal(resp.Result, &data); err != nil {
				result.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    w.baseWrapper.GetHostName(),
					ServiceName: w.serviceName,
				}
			} else {
				result.Items = data
			}
		}

		respCh <- result
	}()

	return respCh
}

func (w *Wrapper) GetTokenomicsStatsByUserId(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetTokenomicsStatsByUserIdResponseChan {
	respCh := make(chan GetTokenomicsStatsByUserIdResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "GetTokenomicsStatsByUserId", GetTokenomicsStatsByUserIdRequest{
		UserIds: userIds,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	go func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := GetTokenomicsStatsByUserIdResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			var data = make(map[int64]*UserTokenomicsStats)

			if err := json.Unmarshal(resp.Result, &data); err != nil {
				result.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    w.baseWrapper.GetHostName(),
					ServiceName: w.serviceName,
				}
			} else {
				result.Items = data
			}
		}

		respCh <- result
	}()

	return respCh
}

func (w *Wrapper) GetConfigProperties(properties []string, apmTransaction *apm.Transaction, forceLog bool) chan GetConfigPropertiesResponseChan {
	respCh := make(chan GetConfigPropertiesResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "GetConfigProperties", GetConfigPropertiesRequest{Properties: properties},
		map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	go func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := GetConfigPropertiesResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			var data = make(map[string]string)

			if err := json.Unmarshal(resp.Result, &data); err != nil {
				result.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    w.baseWrapper.GetHostName(),
					ServiceName: w.serviceName,
				}
			} else {
				result.Items = data
			}
		}
		respCh <- result
	}()

	return respCh
}

func (w *Wrapper) GetReferralsInfo(referrerId int64, referralIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[GetReferralInfoResponse] {
	return wrappers.ExecuteRpcRequestAsync[GetReferralInfoResponse](w.baseWrapper, w.apiUrl, "GetReferralsInfo", GetReferralInfoRequest{
		ReferralIds: referralIds,
		ReferrerId:  referrerId,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)
}

func (w *Wrapper) GetActivitiesInfo(userId int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[GetActivitiesInfoResponse] {
	return wrappers.ExecuteRpcRequestAsync[GetActivitiesInfoResponse](w.baseWrapper, w.apiUrl, "GetActivitiesInfo", GetActivitiesInfoRequest{
		UserId: userId,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)
}

func (w *Wrapper) CreateBotViews(botViews map[int64][]int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[any] {
	return wrappers.ExecuteRpcRequestAsync[any](w.baseWrapper, w.apiUrl, "CreateBotViews", CreateBotViewsRequest{
		BotViews: botViews,
	}, map[string]string{}, w.defaultTimeout, apm.TransactionFromContext(ctx), w.serviceName, forceLog)
}

func (w *Wrapper) WriteOffUserTokensForAd(userId int64, adCampaignId int64, amount decimal.Decimal, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[any] {
	return wrappers.ExecuteRpcRequestAsync[any](w.baseWrapper, w.apiUrl, "WriteOffUserTokensForAd", WriteOffUserTokensForAdRequest{
		UserId:       userId,
		AdCampaignId: adCampaignId,
		Amount:       amount,
	}, map[string]string{}, w.defaultTimeout, apm.TransactionFromContext(ctx), w.serviceName, forceLog)
}

func (w *Wrapper) GetReferralsProgressInfo(referrerId int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[GetReferralProgressInfoResponse] {
	return wrappers.ExecuteRpcRequestAsync[GetReferralProgressInfoResponse](w.baseWrapper, w.apiUrl, "GetReferralsProgressInfo", GetReferralProgressInfoRequest{
		ReferrerId: referrerId,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)
}

func (w *Wrapper) GetMyReferredUsersWatchedVideoInfo(referrerId, page, count int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[GetMyReferredUsersWatchedVideoInfoResponse] {
	return wrappers.ExecuteRpcRequestAsync[GetMyReferredUsersWatchedVideoInfoResponse](w.baseWrapper, w.apiUrl, "GetMyReferredUsersWatchedVideoInfo", GetMyReferredUsersWatchedVideoInfoRequest{
		ReferrerId: referrerId,
		Count:      count,
		Page:       page,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)
}

func (w *Wrapper) DeductVaultPointsForIntroFeed(userId int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[DeductVaultPointsForIntroFeedResponse] {
	return wrappers.ExecuteRpcRequestAsync[DeductVaultPointsForIntroFeedResponse](w.baseWrapper, w.apiUrl, "DeductVaultPointsForIntroFeed", DeductVaultPointsForIntroFeedRequest{
		UserId: userId,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)
}

func (w *Wrapper) AddPointsToVault(userId int64, points decimal.Decimal, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[AddPointsToVaultResponse] {
	return wrappers.ExecuteRpcRequestAsync[AddPointsToVaultResponse](w.baseWrapper, w.apiUrl, "AddPointsToVault", AddPointsToVaultRequest{
		UserId: userId,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)
}
