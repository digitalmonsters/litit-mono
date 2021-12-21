package user_hashtag

import "github.com/digitalmonsters/go-common/rpc"

//goland:noinspection GoNameStartsWithPackageName
type GetUserHashtagSubscriptionStateResponseChan struct {
	Error *rpc.RpcError  `json:"error"`
	Data  map[int64]bool `json:"data"`
}

//goland:noinspection GoNameStartsWithPackageName
type GetUserHashtagSubscriptionStateBulkRequest struct {
	UserId   int64    `json:"user_id"`
	Hashtags []string `json:"hashtags"`
}
