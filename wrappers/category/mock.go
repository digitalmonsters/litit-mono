package category

import (
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
)

type CategoryWrapperMock struct {
	GetCategoryInternalFn func(categoryIds []int64, omitCategoryIds []int64, limit int, offset int, onlyParent null.Bool, apmTransaction *apm.Transaction, forceLog bool) chan CategoryGetInternalResponseChan
}

func (w *CategoryWrapperMock) GetCategoryInternal(categoryIds []int64, omitCategoryIds []int64, limit int, offset int, onlyParent null.Bool, apmTransaction *apm.Transaction, forceLog bool) chan CategoryGetInternalResponseChan {
	return w.GetCategoryInternalFn(categoryIds, omitCategoryIds, limit, offset, onlyParent, apmTransaction, forceLog)
}

func GetMock() ICategoryWrapper {
	return &CategoryWrapperMock{}
}
