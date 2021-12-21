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
