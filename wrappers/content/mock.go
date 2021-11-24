package content

import "go.elastic.co/apm"

type ContentWrapperMock struct {
	GetInternalFn func(contentIds []int64, includeDeleted bool, apmTransaction *apm.Transaction, forceLog bool) chan ContentGetInternalResponseChan
}

func (w *ContentWrapperMock) GetInternal(contentIds []int64, includeDeleted bool, apmTransaction *apm.Transaction, forceLog bool) chan ContentGetInternalResponseChan {
	return w.GetInternalFn(contentIds, includeDeleted, apmTransaction, forceLog)
}

func GetMock() IContentWrapper { // for compiler errors
	return &ContentWrapperMock{}
}
