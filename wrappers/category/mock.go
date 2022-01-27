package category

import (
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
)

type CategoryWrapperMock struct {
	GetCategoryInternalFn func(categoryIds []int64, omitCategoryIds []int64, limit int, offset int, onlyParent null.Bool, withViews null.Bool, apmTransaction *apm.Transaction,
		shouldHaveValidContent bool, forceLog bool) chan CategoryGetInternalResponseChan
	GetAllCategoriesFn             func(categoryIds []int64, includeDeleted bool, apmTransaction *apm.Transaction, forceLog bool) chan GetAllCategoriesResponseChan
	GetUserBlacklistedCategoriesFn func(userId int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUserBlacklistedCategoriesChan
}

func (w *CategoryWrapperMock) GetCategoryInternal(categoryIds []int64, omitCategoryIds []int64, limit int, offset int, onlyParent null.Bool, withViews null.Bool,
	apmTransaction *apm.Transaction, shouldHaveValidContent bool, forceLog bool) chan CategoryGetInternalResponseChan {
	return w.GetCategoryInternalFn(categoryIds, omitCategoryIds, limit, offset, onlyParent, withViews, apmTransaction,
		shouldHaveValidContent, forceLog)
}

func (w *CategoryWrapperMock) GetAllCategories(categoryIds []int64, includeDeleted bool, apmTransaction *apm.Transaction, forceLog bool) chan GetAllCategoriesResponseChan {
	return w.GetAllCategoriesFn(categoryIds, includeDeleted, apmTransaction, forceLog)
}

func (w *CategoryWrapperMock) GetUserBlacklistedCategories(userId int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUserBlacklistedCategoriesChan {
	return w.GetUserBlacklistedCategoriesFn(userId, apmTransaction, forceLog)
}

func GetMock() ICategoryWrapper {
	return &CategoryWrapperMock{}
}
