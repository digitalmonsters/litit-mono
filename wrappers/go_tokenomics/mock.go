package go_tokenomics

import (
	"github.com/digitalmonsters/go-common/wrappers"
	"go.elastic.co/apm"
)

type GoTokenomicsWrapperMock struct {
	GetUsersTokenomicsInfoFn              func(userIds []int64, filters []Filter, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersTokenomicsInfoResponseChan
	GetWithdrawalsAmountsByAdminIdsFn     func(adminIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetWithdrawalsAmountsByAdminIdsResponseChan
	GetContentEarningsTotalByContentIdsFn func(contentIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetContentEarningsTotalByContentIdsResponseChan
	GetTokenomicsStatsByUserIdFn          func(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetTokenomicsStatsByUserIdResponseChan
	GetConfigPropertiesFn                 func(properties []string, apmTransaction *apm.Transaction, forceLog bool) chan GetConfigPropertiesResponseChan
	GetReferralsInfoFn                    func(referrerId int64, referralIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[GetReferralInfoResponse]
}

func (w *GoTokenomicsWrapperMock) GetUsersTokenomicsInfo(userIds []int64, filters []Filter, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersTokenomicsInfoResponseChan {
	return w.GetUsersTokenomicsInfoFn(userIds, filters, apmTransaction, forceLog)
}

func (w *GoTokenomicsWrapperMock) GetWithdrawalsAmountsByAdminIds(adminIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetWithdrawalsAmountsByAdminIdsResponseChan {
	return w.GetWithdrawalsAmountsByAdminIdsFn(adminIds, apmTransaction, forceLog)
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

func GetMock() IGoTokenomicsWrapper {
	return &GoTokenomicsWrapperMock{}
}
