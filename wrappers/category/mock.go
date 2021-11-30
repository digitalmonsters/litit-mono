package category

import (
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
)

type CategoryWrapperMock struct {
	GetCategoryInternalFn func(categoryIds []int64, limit int, offset int, userId null.Int, excludeRoot bool, excludeFollowing bool, apmTransaction *apm.Transaction, forceLog bool) chan CategoryGetInternalResponseChan
}

func (w *CategoryWrapperMock) GetCategoryInternal(categoryIds []int64, limit int, offset int, userId null.Int, excludeRoot bool, excludeFollowing bool, apmTransaction *apm.Transaction, forceLog bool) chan CategoryGetInternalResponseChan {
	return w.GetCategoryInternalFn(categoryIds, limit, offset, userId, excludeRoot, excludeFollowing, apmTransaction, forceLog)
}

func GetMock() ICategoryWrapper {
	return &CategoryWrapperMock{}
}
