package category

import (
	"github.com/digitalmonsters/go-common/rpc"
	"gopkg.in/guregu/null.v4"
)

type SimpleCategory struct {
	Id         int64  `json:"id"`
	Name       string `json:"name"`
	ViewsCount int    `json:"views_count"`
	Emojis     string `json:"emojis"`
}

type Status int

const (
	StatusNotActive  Status = 0
	StatusActive     Status = 1
	StatusComingSoon Status = 2
)

type ResponseData struct {
	Items      []SimpleCategory `json:"items"`
	TotalCount int64            `json:"total_count"`
}

type CategoryGetInternalResponseChan struct {
	Error *rpc.RpcError `json:"error"`
	Data  *ResponseData `json:"data"`
}

type GetUserBlacklistedCategoriesRequest struct {
	UserId int64 `json:"user_id"`
}

type GetUserBlacklistedCategoriesChan struct {
	Error *rpc.RpcError                         `json:"error"`
	Data  *GetUserBlacklistedCategoriesResponse `json:"data"`
}

type GetUserBlacklistedCategoriesResponse struct {
	CategoryIds []int64 `json:"category_ids"`
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

type GetAllCategoriesResponseChan struct {
	Error *rpc.RpcError                       `json:"error"`
	Data  map[int64]AllCategoriesResponseItem `json:"data"`
}

type AllCategoriesResponseItem struct {
	Id        int64    `json:"id"`
	Name      string   `json:"name"`
	Emojis    string   `json:"emojis"`
	ParentId  null.Int `json:"parent_id"`
	SortOrder int      `json:"sort_order"`
	Status    Status   `json:"status"`
}
