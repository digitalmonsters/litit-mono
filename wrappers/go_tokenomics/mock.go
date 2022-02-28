package go_tokenomics

import "go.elastic.co/apm"

type GoTokenomicsWrapperMock struct {
	GetUsersTokenomicsInfoFn              func(userIds []int64, filters []Filter, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersTokenomicsInfoResponseChan
	GetWithdrawalsAmountsByAdminIdsFn     func(adminIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetWithdrawalsAmountsByAdminIdsResponseChan
	GetContentEarningsTotalByContentIdsFn func(contentIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetContentEarningsTotalByContentIdsResponseChan
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

func GetMock() IGoTokenomicsWrapper {
	return &GoTokenomicsWrapperMock{}
}
