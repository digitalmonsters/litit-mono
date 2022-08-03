package user_category

import "github.com/digitalmonsters/go-common/rpc"

//goland:noinspection GoNameStartsWithPackageName
type GetUserCategorySubscriptionStateResponseChan struct {
	Error *rpc.RpcError  `json:"error"`
	Data  map[int64]bool `json:"data"`
}

//goland:noinspection GoNameStartsWithPackageName
type GetUserCategorySubscriptionStateBulkRequest struct {
	UserId      int64   `json:"user_id"`
	CategoryIds []int64 `json:"category_ids"`
}

type GetInternalUserCategorySubscriptionsRequest struct {
	UserId    int64  `json:"user_id"`
	Limit     int    `json:"limit"`
	PageState string `json:"page_state"`
}

type GetInternalUserCategorySubscriptionsResponse struct {
	CategoryIds []int64 `json:"category_ids"`
	PageState   string  `json:"page_state"`
}
