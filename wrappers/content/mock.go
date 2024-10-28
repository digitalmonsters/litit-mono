package content

import (
	"context"

	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"

	"github.com/digitalmonsters/go-common/frontend"
	"github.com/digitalmonsters/go-common/wrappers"
)

type ContentWrapperMock struct {
	GetInternalFn              func(contentIds []int64, includeDeleted bool, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[int64]SimpleContent]
	GetContentIdListInternalFn func(pageNo int64, pageSize int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[ContentListIdPaginationResponse]
	GetInternalAdminModelsFn   func(contentIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[int64]frontend.ContentModel]
	GetTopNotFollowingUsersFn  func(userId int64, limit int, offset int, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[GetTopNotFollowingUsersResponse]
	GetHashtagsInternalFn      func(hashtags []string, omitHashtags []string, limit int, offset int, withViews null.Bool, apmTransaction *apm.Transaction,
		shouldHaveValidContent bool, forceLog bool) chan wrappers.GenericResponseChan[HashtagResponseData]

	GetCategoryInternalFn func(categoryIds []int64, omitCategoryIds []int64, limit int, offset int, onlyParent null.Bool, withViews null.Bool, apmTransaction *apm.Transaction,
		shouldHaveValidContent bool, forceLog bool) chan wrappers.GenericResponseChan[CategoryResponseData]
	GetAllCategoriesFn             func(categoryIds []int64, includeDeleted bool, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[int64]AllCategoriesResponseItem]
	GetUserBlacklistedCategoriesFn func(userId int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[GetUserBlacklistedCategoriesResponse]
	GetUserLikesFn                 func(userId int64, limit int, offset int, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[LikedContent]
	GetConfigPropertiesFn          func(properties []string, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[string]string]
	GetRejectReasonFn              func(ids []int64, includeDeleted bool, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]RejectReason]
	GetTopUsersInCategoriesFn      func(ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64][]int64]
	InsertMusicContentFn           func(content MusicContentRequest, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[SimpleContent]
	GetLastContentFn               func(ctx context.Context, userId int64) chan wrappers.GenericResponseChan[[]SimpleContent]
}

func (w *ContentWrapperMock) GetLastContent(ctx context.Context, userId int64) chan wrappers.GenericResponseChan[[]SimpleContent] {
	return w.GetLastContentFn(ctx, userId)
}

func (w *ContentWrapperMock) GetInternal(contentIds []int64, includeDeleted bool, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[int64]SimpleContent] {
	return w.GetInternalFn(contentIds, includeDeleted, apmTransaction, forceLog)
}

func (w *ContentWrapperMock) GetContentIdListInternal(pageNo int64, pageSize int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[ContentListIdPaginationResponse] {
	return w.GetContentIdListInternalFn(pageNo, pageSize, apmTransaction, forceLog)
}

func (w *ContentWrapperMock) GetInternalAdminModels(contentIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[int64]frontend.ContentModel] {
	return w.GetInternalAdminModelsFn(contentIds, apmTransaction, forceLog)
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
func (w *ContentWrapperMock) GetConfigProperties(properties []string, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[string]string] {
	return w.GetConfigPropertiesFn(properties, apmTransaction, forceLog)
}

func (w *ContentWrapperMock) GetRejectReason(ids []int64, includeDeleted bool, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]RejectReason] {
	return w.GetRejectReasonFn(ids, includeDeleted, ctx, forceLog)
}

func (w *ContentWrapperMock) GetTopUsersInCategories(ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64][]int64] {
	return w.GetTopUsersInCategoriesFn(ctx, forceLog)
}

func (w *ContentWrapperMock) InsertMusicContent(content MusicContentRequest, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[SimpleContent] {
	return w.InsertMusicContentFn(content, ctx, forceLog)
}

func (w *ContentWrapperMock) GetIfIntroExists(ctx context.Context, userId int64) chan wrappers.GenericResponseChan[IntroExists] {
	panic("implement me")
}

func GetMock() IContentWrapper { // for compiler errors
	return &ContentWrapperMock{}
}
