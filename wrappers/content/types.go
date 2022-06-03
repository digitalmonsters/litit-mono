package content

import (
	"github.com/digitalmonsters/go-common/eventsourcing"
	"gopkg.in/guregu/null.v4"
	"time"
)

type SimpleContent struct {
	Id            int64                     `json:"id"`
	ContentType   eventsourcing.ContentType `json:"type"`
	Duration      int                       `json:"duration"`
	AgeRestricted bool                      `json:"age_restricted"`
	AuthorId      int64                     `json:"author_id"`
	CategoryId    null.Int                  `json:"category_id"`
	Hashtags      []string                  `json:"hashtags"`
	Width         int                       `json:"width"`
	Height        int                       `json:"height"`
	VideoId       string                    `json:"video_id"`
	SubCategoryId null.Int                  `json:"sub_category_id"`
	Unlisted      bool                      `json:"unlisted"`
	Draft         bool                      `json:"draft"`
	Deleted       bool                      `json:"deleted"`
	AllowComments bool                      `json:"allow_comments"`
	Approved      bool                      `json:"approved"`
}

type GetTopNotFollowingUsersRequest struct {
	UserId int64 `json:"user_id"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}

type GetTopNotFollowingUsersResponse struct {
	Items      []int64 `json:"items"`
	TotalCount int64   `json:"total_count"`
}

type ContentGetInternalRequest struct {
	IncludeDeleted bool    `json:"include_deleted"`
	ContentIds     []int64 `json:"content_ids"`
}

type HashtagResponseData struct {
	Items      []SimpleHashtagModel `json:"items"`
	TotalCount int64                `json:"total_count"`
}

type SimpleHashtagModel struct {
	Name       string `json:"name"`
	ViewsCount int    `json:"views_count"`
}

type GetHashtagsInternalRequest struct {
	Hashtags               []string  `json:"hashtags"`
	OmitHashtags           []string  `json:"omit_hashtags"`
	WithViews              null.Bool `json:"with_views"`
	ShouldHaveValidContent bool      `json:"should_have_valid_content"`
	Limit                  int       `json:"limit"`
	Offset                 int       `json:"offset"`
}

type CategoryResponseData struct {
	Items      []SimpleCategoryModel `json:"items"`
	TotalCount int64                 `json:"total_count"`
}

type SimpleCategoryModel struct {
	Id         int64  `json:"id"`
	Name       string `json:"name"`
	ViewsCount int    `json:"views_count"`
	Emojis     string `json:"emojis"`
}

type AllCategoriesResponseItem struct {
	Id        int64          `json:"id"`
	Name      string         `json:"name"`
	Emojis    string         `json:"emojis"`
	ParentId  null.Int       `json:"parent_id"`
	SortOrder int            `json:"sort_order"`
	Status    CategoryStatus `json:"status"`
}

type GetUserBlacklistedCategoriesResponse struct {
	CategoryIds []int64 `json:"category_ids"`
}

type CategoryStatus int

const (
	CategoryStatusNotActive  CategoryStatus = 0
	CategoryStatusActive     CategoryStatus = 1
	CategoryStatusComingSoon CategoryStatus = 2
)

type GetUserBlacklistedCategoriesRequest struct {
	UserId int64 `json:"user_id"`
}

type GetCategoryInternalRequest struct {
	CategoryIds            []int64   `json:"category_ids"`
	Limit                  int       `json:"limit"`
	Offset                 int       `json:"offset"`
	OnlyParent             null.Bool `json:"only_parent"`
	WithViews              null.Bool `json:"with_views"`
	ShouldHaveValidContent bool      `json:"should_have_valid_content"`
	OmitCategoryIds        []int64   `json:"omit_category_ids"`
}

type GetAllCategoriesRequest struct {
	CategoryIds    []int64 `json:"category_ids"`
	IncludeDeleted bool    `json:"include_deleted"`
}

type LikedContent struct {
	ContentIds []int64 `json:"content_ids"`
	TotalCount int64   `json:"total_count"`
}

type GetUserLikesRequest struct {
	UserId int64 `json:"user_id"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}

type GetConfigValuesRequest struct {
	Properties []string `json:"properties"`
}

type RejectReasonType int

const (
	ReasonTypeNone  = RejectReasonType(0)
	ReasonTypeVideo = RejectReasonType(1)
	ReasonTypeSpot  = RejectReasonType(2)
)

type RejectReason struct {
	Id        int64            `json:"id"`
	Type      RejectReasonType `json:"type"`
	Reason    string           `json:"reason"`
	Active    bool             `json:"active"`
	CreatedAt time.Time        `json:"created_at"`
	DeletedAt null.Time        `json:"deleted_at"`
}

type GetContentRejectReasonRequest struct {
	Ids            []int64 `json:"ids"`
	IncludeDeleted bool    `json:"include_deleted"`
}
