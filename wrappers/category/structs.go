package category

import "github.com/digitalmonsters/go-common/rpc"

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

type GetCategoryInternalRequest struct {
	CategoryIds []int64 `json:"category_ids"`
	Limit       int     `json:"limit"`
	Offset      int     `json:"offset"`
	ExcludeRoot bool    `json:"exclude_root"`
}
