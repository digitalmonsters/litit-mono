package content

import (
	"github.com/digitalmonsters/go-common/wrappers"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
)

type ContentWrapperMock struct {
	GetInternalFn             func(contentIds []int64, includeDeleted bool, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[int64]SimpleContent]
	GetTopNotFollowingUsersFn func(userId int64, limit int, offset int, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[GetTopNotFollowingUsersResponse]
	GetHashtagsInternalFn     func(hashtags []string, omitHashtags []string, limit int, offset int, withViews null.Bool, apmTransaction *apm.Transaction,
		shouldHaveValidContent bool, forceLog bool) chan wrappers.GenericResponseChan[HashtagResponseData]

	GetCategoryInternalFn func(categoryIds []int64, omitCategoryIds []int64, limit int, offset int, onlyParent null.Bool, withViews null.Bool, apmTransaction *apm.Transaction,
		shouldHaveValidContent bool, forceLog bool) chan wrappers.GenericResponseChan[CategoryResponseData]
	GetAllCategoriesFn             func(categoryIds []int64, includeDeleted bool, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[int64]AllCategoriesResponseItem]
	GetUserBlacklistedCategoriesFn func(userId int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[GetUserBlacklistedCategoriesResponse]
	GetUserLikesFn                 func(userId int64, limit int, offset int, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[LikedContent]
}

func (w *ContentWrapperMock) GetInternal(contentIds []int64, includeDeleted bool, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[int64]SimpleContent] {
	return w.GetInternalFn(contentIds, includeDeleted, apmTransaction, forceLog)
}

func (w *ContentWrapperMock) GetTopNotFollowingUsers(userId int64, limit int, offset int, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[GetTopNotFollowingUsersResponse] {
	return w.GetTopNotFollowingUsersFn(userId, limit, offset, apmTransaction, forceLog)
}

func (w *ContentWrapperMock) GetHashtagsInternal(hashtags []string, omitHashtags []string, limit int, offset int,
	withViews null.Bool, apmTransaction *apm.Transaction, shouldHaveValidContent bool, forceLog bool) chan wrappers.GenericResponseChan[HashtagResponseData] {
	return w.GetHashtagsInternalFn(hashtags, omitHashtags, limit, offset, withViews, apmTransaction, shouldHaveValidContent, forceLog)
}

func (w *ContentWrapperMock) GetCategoryInternal(categoryIds []int64, omitCategoryIds []int64, limit int, offset int, onlyParent null.Bool, withViews null.Bool,
	apmTransaction *apm.Transaction, shouldHaveValidContent bool, forceLog bool) chan wrappers.GenericResponseChan[CategoryResponseData] {
	return w.GetCategoryInternalFn(categoryIds, omitCategoryIds, limit, offset, onlyParent, withViews, apmTransaction,
		shouldHaveValidContent, forceLog)
}

func (w *ContentWrapperMock) GetAllCategories(categoryIds []int64, includeDeleted bool, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[int64]AllCategoriesResponseItem] {
	return w.GetAllCategoriesFn(categoryIds, includeDeleted, apmTransaction, forceLog)
}

func (w *ContentWrapperMock) GetUserBlacklistedCategories(userId int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[GetUserBlacklistedCategoriesResponse] {
	return w.GetUserBlacklistedCategoriesFn(userId, apmTransaction, forceLog)
}

func (w *ContentWrapperMock) GetUserLikes(userId int64, limit int, offset int, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[LikedContent] {
	return w.GetUserLikesFn(userId, limit, offset, apmTransaction, forceLog)
}

func GetMock() IContentWrapper { // for compiler errors
	return &ContentWrapperMock{}
}
