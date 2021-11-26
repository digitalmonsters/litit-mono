package category

import "go.elastic.co/apm"

type CategoryWrapperMock struct {
	GetCategoryInternalFn func(categoryIds []int64, limit int64, offset int64, apmTransaction *apm.Transaction, forceLog bool) chan CategoryGetInternalResponseChan
}

func (w *CategoryWrapperMock) GetCategoryInternal(categoryIds []int64, limit int64, offset int64, apmTransaction *apm.Transaction, forceLog bool) chan CategoryGetInternalResponseChan {
	return w.GetCategoryInternalFn(categoryIds, limit, offset, apmTransaction, forceLog)
}

func GetMock() ICategoryWrapper {
	return &CategoryWrapperMock{}
}
