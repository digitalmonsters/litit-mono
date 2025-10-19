package go_tokenomics

import (
	"context"

	"github.com/digitalmonsters/go-common/filters"
	"github.com/digitalmonsters/go-common/wrappers"
	"github.com/shopspring/decimal"
	"go.elastic.co/apm"
)

type GoTokenomicsWrapperMock struct {
	GetUsersTokenomicsInfoFn              func(userIds []int64, filters []filters.Filter, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]UserTokenomicsInfo]
	GetWithdrawalsAmountsByAdminIdsFn     func(adminIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]decimal.Decimal]
	GetContentEarningsTotalByContentIdsFn func(contentIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetContentEarningsTotalByContentIdsResponseChan
	GetTokenomicsStatsByUserIdFn          func(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetTokenomicsStatsByUserIdResponseChan
	GetConfigPropertiesFn                 func(properties []string, apmTransaction *apm.Transaction, forceLog bool) chan GetConfigPropertiesResponseChan
	GetReferralsInfoFn                    func(referrerId int64, referralIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[GetReferralInfoResponse]
	GetActivitiesInfoFn                   func(userId int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[GetActivitiesInfoResponse]
	CreateBotViewsFn                      func(botViews map[int64][]int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[any]
	WriteOffUserTokensForAdFn             func(userId int64, adCampaignId int64, amount decimal.Decimal, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[any]
	GetReferralsProgressInfoFn            func(referrerId int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[GetReferralProgressInfoResponse]
	GetMyReferredUsersWatchedVideoInfoFn  func(referrerId, page, count int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[GetMyReferredUsersWatchedVideoInfoResponse]
	DeductVaultPointsForIntroFeedFn       func(userId int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[DeductVaultPointsForIntroFeedResponse]
	AddPointsToVaultFn                    func(userId int64, points decimal.Decimal, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[AddPointsToVaultResponse]
}

func (w *GoTokenomicsWrapperMock) GetUsersTokenomicsInfo(userIds []int64, filters []filters.Filter, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]UserTokenomicsInfo] {
	return w.GetUsersTokenomicsInfoFn(userIds, filters, ctx, forceLog)
}

func (w *GoTokenomicsWrapperMock) GetWithdrawalsAmountsByAdminIds(adminIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]decimal.Decimal] {
	return w.GetWithdrawalsAmountsByAdminIdsFn(adminIds, ctx, forceLog)
}

func (w *GoTokenomicsWrapperMock) GetContentEarningsTotalByContentIds(contentIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetContentEarningsTotalByContentIdsResponseChan {
	return w.GetContentEarningsTotalByContentIdsFn(contentIds, apmTransaction, forceLog)
}

func (w *GoTokenomicsWrapperMock) GetTokenomicsStatsByUserId(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetTokenomicsStatsByUserIdResponseChan {
	return w.GetTokenomicsStatsByUserIdFn(userIds, apmTransaction, forceLog)
}

func (w *GoTokenomicsWrapperMock) GetConfigProperties(properties []string, apmTransaction *apm.Transaction, forceLog bool) chan GetConfigPropertiesResponseChan {
	return w.GetConfigPropertiesFn(properties, apmTransaction, forceLog)
}
func (w *GoTokenomicsWrapperMock) GetReferralsInfo(referrerId int64, referralIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[GetReferralInfoResponse] {
	return w.GetReferralsInfoFn(referrerId, referralIds, apmTransaction, forceLog)
}
func (w *GoTokenomicsWrapperMock) GetActivitiesInfo(userId int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[GetActivitiesInfoResponse] {
	return w.GetActivitiesInfoFn(userId, apmTransaction, forceLog)
}

func (w *GoTokenomicsWrapperMock) CreateBotViews(botViews map[int64][]int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[any] {
	return w.CreateBotViewsFn(botViews, ctx, forceLog)
}

func (w *GoTokenomicsWrapperMock) WriteOffUserTokensForAd(userId int64, adCampaignId int64, amount decimal.Decimal, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[any] {
	return w.WriteOffUserTokensForAdFn(userId, adCampaignId, amount, ctx, forceLog)
}

func (w *GoTokenomicsWrapperMock) GetReferralsProgressInfo(referrerId int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[GetReferralProgressInfoResponse] {
	return w.GetReferralsProgressInfoFn(referrerId, apmTransaction, forceLog)
}

func (w *GoTokenomicsWrapperMock) GetMyReferredUsersWatchedVideoInfo(referrerId, page, count int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[GetMyReferredUsersWatchedVideoInfoResponse] {
	return w.GetMyReferredUsersWatchedVideoInfoFn(referrerId, page, count, apmTransaction, forceLog)
}

func (w *GoTokenomicsWrapperMock) DeductVaultPointsForIntroFeed(userId int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[DeductVaultPointsForIntroFeedResponse] {
	return w.DeductVaultPointsForIntroFeedFn(userId, apmTransaction, forceLog)
}

func (w *GoTokenomicsWrapperMock) AddPointsToVault(userId int64, points decimal.Decimal, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[AddPointsToVaultResponse] {
	return w.AddPointsToVaultFn(userId, points, apmTransaction, forceLog)
}

func GetMock() IGoTokenomicsWrapper {
	return &GoTokenomicsWrapperMock{}
}
