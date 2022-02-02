package go_tokenomics

import "go.elastic.co/apm"

type GoTokenomicsWrapperMock struct {
	GetUsersTokenomicsInfoFn func(userIds []int64, filters []Filter, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersTokenomicsInfoResponseChan
}

func (w *GoTokenomicsWrapperMock) GetUsersTokenomicsInfo(userIds []int64, filters []Filter, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersTokenomicsInfoResponseChan {
	return w.GetUsersTokenomicsInfoFn(userIds, filters, apmTransaction, forceLog)
}

func GetMock() IGoTokenomicsWrapper {
	return &GoTokenomicsWrapperMock{}
}
