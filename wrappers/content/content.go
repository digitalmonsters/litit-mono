package content

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"

	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/frontend"
	"github.com/digitalmonsters/go-common/wrappers"
)

type IContentWrapper interface {
	GetInternal(contentIds []int64, includeDeleted bool, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[int64]SimpleContent]
	GetContentIdListInternal(pageNo int64, pageSize int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[ContentListIdPaginationResponse]
	GetInternalAdminModels(contentIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[int64]frontend.ContentModel]
	GetTopNotFollowingUsers(userId int64, limit int, offset int, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[GetTopNotFollowingUsersResponse]
	GetHashtagsInternal(hashtags []string, omitHashtags []string, limit int, offset int, withViews null.Bool, apmTransaction *apm.Transaction,
		shouldHaveValidContent bool, forceLog bool) chan wrappers.GenericResponseChan[HashtagResponseData]

	GetCategoryInternal(categoryIds []int64, omitCategoryIds []int64, limit int, offset int, onlyParent null.Bool, withViews null.Bool,
		apmTransaction *apm.Transaction, shouldHaveValidContent bool, forceLog bool) chan wrappers.GenericResponseChan[CategoryResponseData]
	GetAllCategories(categoryIds []int64, includeDeleted bool, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[int64]AllCategoriesResponseItem]
	GetUserBlacklistedCategories(userId int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[GetUserBlacklistedCategoriesResponse]
	GetUserLikes(userId int64, limit int, offset int, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[LikedContent]
	GetConfigProperties(properties []string, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[string]string]
	GetRejectReason(ids []int64, includeDeleted bool, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]RejectReason]
	GetTopUsersInCategories(ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64][]int64]
	InsertMusicContent(content MusicContentRequest, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[SimpleContent]
	GetLastContent(ctx context.Context, userId int64) chan wrappers.GenericResponseChan[[]SimpleContent]
	GetIfIntroExists(ctx context.Context, userId []int64) chan wrappers.GenericResponseChan[[]IntroExists]
	GetAllUploadCount(ctx context.Context, userId int64) chan wrappers.GenericResponseChan[UploadCountResponse]
}

//goland:noinspection GoNameStartsWithPackageName
type ContentWrapper struct {
	baseWrapper    *wrappers.BaseWrapper
	defaultTimeout time.Duration
	apiUrl         string
	serviceName    string
}

func NewContentWrapper(config boilerplate.WrapperConfig) IContentWrapper {
	timeout := 5 * time.Second

	if config.TimeoutSec > 0 {
		timeout = time.Duration(config.TimeoutSec) * time.Second
	}

	if len(config.ApiUrl) == 0 {
		config.ApiUrl = "http://content"

		log.Warn().Msgf("Api Url is missing for Content. Setting as default : %v", config.ApiUrl)
	}

	return &ContentWrapper{
		baseWrapper:    wrappers.GetBaseWrapper(),
		defaultTimeout: timeout,
		apiUrl:         fmt.Sprintf("%v/rpc-service", common.StripSlashFromUrl(config.ApiUrl)),
		serviceName:    "content",
	}
}

func (w *ContentWrapper) GetLastContent(ctx context.Context, userId int64) chan wrappers.GenericResponseChan[[]SimpleContent] {
	return wrappers.ExecuteRpcRequestAsync[[]SimpleContent](w.baseWrapper, w.apiUrl, "GetLastContentInternal", GetLastContentRequest{
		UserId: userId,
	}, map[string]string{}, w.defaultTimeout, apm.TransactionFromContext(ctx), w.serviceName, false)
}

func (w *ContentWrapper) GetInternal(contentIds []int64, includeDeleted bool, apmTransaction *apm.Transaction,
	forceLog bool) chan wrappers.GenericResponseChan[map[int64]SimpleContent] {
	return wrappers.ExecuteRpcRequestAsync[map[int64]SimpleContent](w.baseWrapper, w.apiUrl, "ContentGetInternal", ContentGetInternalRequest{
		ContentIds:     contentIds,
		IncludeDeleted: includeDeleted,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)
}

func (w *ContentWrapper) GetContentIdListInternal(pageNo int64, pageSize int64, apmTransaction *apm.Transaction,
	forceLog bool) chan wrappers.GenericResponseChan[ContentListIdPaginationResponse] {
	return wrappers.ExecuteRpcRequestAsync[ContentListIdPaginationResponse](w.baseWrapper, w.apiUrl, "ListContentIdInternal", ContentListIdPagination{
		PageNo:   pageNo,
		PageSize: pageSize,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)
}

func (w *ContentWrapper) GetInternalAdminModels(contentIds []int64, apmTransaction *apm.Transaction,
	forceLog bool) chan wrappers.GenericResponseChan[map[int64]frontend.ContentModel] {
	return wrappers.ExecuteRpcRequestAsync[map[int64]frontend.ContentModel](w.baseWrapper, w.apiUrl, "ContentGetInternalAdminModels", ContentGetInternalAdminModelsRequest{
		ContentIds: contentIds,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)
}

func (w *ContentWrapper) GetTopNotFollowingUsers(userId int64, limit int, offset int, apmTransaction *apm.Transaction,
	forceLog bool) chan wrappers.GenericResponseChan[GetTopNotFollowingUsersResponse] {

	return wrappers.ExecuteRpcRequestAsync[GetTopNotFollowingUsersResponse](w.baseWrapper, w.apiUrl, "GetTopNotFollowingUsers", GetTopNotFollowingUsersRequest{
		UserId: userId,
		Limit:  limit,
		Offset: offset,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)
}

func (w *ContentWrapper) GetHashtagsInternal(hashtags []string, omitHashtags []string, limit int, offset int, withViews null.Bool,
	apmTransaction *apm.Transaction, shouldHaveValidContent bool, forceLog bool) chan wrappers.GenericResponseChan[HashtagResponseData] {
	return wrappers.ExecuteRpcRequestAsync[HashtagResponseData](w.baseWrapper, w.apiUrl, "GetHashtagsInternal", GetHashtagsInternalRequest{
		Hashtags:               hashtags,
		OmitHashtags:           omitHashtags,
		Limit:                  limit,
		WithViews:              withViews,
		Offset:                 offset,
		ShouldHaveValidContent: shouldHaveValidContent,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)
}

func (w *ContentWrapper) GetUserBlacklistedCategories(userId int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[GetUserBlacklistedCategoriesResponse] {
	return wrappers.ExecuteRpcRequestAsync[GetUserBlacklistedCategoriesResponse](w.baseWrapper, w.apiUrl, "GetUserBlacklistedCategoriesInternal", GetUserBlacklistedCategoriesRequest{
		UserId: userId,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)
}

func (w *ContentWrapper) GetCategoryInternal(categoryIds []int64, omitCategoryIds []int64, limit int, offset int, onlyParent null.Bool, withViews null.Bool,
	apmTransaction *apm.Transaction, shouldHaveValidContent bool, forceLog bool) chan wrappers.GenericResponseChan[CategoryResponseData] {
	return wrappers.ExecuteRpcRequestAsync[CategoryResponseData](w.baseWrapper, w.apiUrl, "GetCategoryInternal", GetCategoryInternalRequest{
		CategoryIds:            categoryIds,
		Limit:                  limit,
		Offset:                 offset,
		OmitCategoryIds:        omitCategoryIds,
		WithViews:              withViews,
		OnlyParent:             onlyParent,
		ShouldHaveValidContent: shouldHaveValidContent,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)
}

func (w *ContentWrapper) GetAllCategories(categoryIds []int64, includeDeleted bool, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[int64]AllCategoriesResponseItem] {
	return wrappers.ExecuteRpcRequestAsync[map[int64]AllCategoriesResponseItem](w.baseWrapper, w.apiUrl, "GetAllCategories", GetAllCategoriesRequest{
		CategoryIds:    categoryIds,
		IncludeDeleted: includeDeleted,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)
}

func (w *ContentWrapper) GetUserLikes(userId int64, limit int, offset int, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[LikedContent] {
	return wrappers.ExecuteRpcRequestAsync[LikedContent](w.baseWrapper, w.apiUrl, "InternalGetUserLikes", GetUserLikesRequest{
		UserId: userId,
		Limit:  limit,
		Offset: offset,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)
}

func (w *ContentWrapper) GetConfigProperties(properties []string, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[string]string] {
	return wrappers.ExecuteRpcRequestAsync[map[string]string](w.baseWrapper, w.apiUrl, "InternalGetConfigValues", GetConfigValuesRequest{Properties: properties},
		map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)
}

func (w *ContentWrapper) GetRejectReason(ids []int64, includeDeleted bool, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]RejectReason] {
	return wrappers.ExecuteRpcRequestAsync[map[int64]RejectReason](w.baseWrapper, w.apiUrl, "InternalGetContentRejectReason", GetContentRejectReasonRequest{
		Ids:            ids,
		IncludeDeleted: includeDeleted,
	}, map[string]string{}, w.defaultTimeout, apm.TransactionFromContext(ctx), w.serviceName, forceLog)
}

func (w *ContentWrapper) GetTopUsersInCategories(ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64][]int64] {
	return wrappers.ExecuteRpcRequestAsync[map[int64][]int64](w.baseWrapper, w.apiUrl, "InternalGetTopUsersInCategories", nil,
		map[string]string{}, w.defaultTimeout, apm.TransactionFromContext(ctx), w.serviceName, forceLog)
}

func (w *ContentWrapper) InsertMusicContent(content MusicContentRequest, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[SimpleContent] {
	return wrappers.ExecuteRpcRequestAsync[SimpleContent](w.baseWrapper, w.apiUrl, "InsertMusicContentInternal", content, map[string]string{}, w.defaultTimeout, apm.TransactionFromContext(ctx), w.serviceName, forceLog)
}

func (w *ContentWrapper) GetIfIntroExists(ctx context.Context, userId []int64) chan wrappers.GenericResponseChan[[]IntroExists] {
	return wrappers.ExecuteRpcRequestAsync[[]IntroExists](w.baseWrapper, w.apiUrl, "GetIfIntroExists", GetIfIntroExistsRequest{
		UserId: userId,
	}, map[string]string{}, w.defaultTimeout, apm.TransactionFromContext(ctx), w.serviceName, false)
}

func (w *ContentWrapper) GetAllUploadCount(ctx context.Context, userId int64) chan wrappers.GenericResponseChan[UploadCountResponse] {
	return wrappers.ExecuteRpcRequestAsync[UploadCountResponse](w.baseWrapper, w.apiUrl, "GetAllUploadCount", GetAllUploadCountRequest{
		UserId: userId,
	}, map[string]string{}, w.defaultTimeout, apm.TransactionFromContext(ctx), w.serviceName, false)
}
