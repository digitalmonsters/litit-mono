package user_likes

import "github.com/digitalmonsters/go-common/rpc"

type GetUserLikesRequest struct {
	UserId int64 `json:"user_id"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}

type LikedContent struct {
	ContentIds []int64 `json:"content_ids"`
	TotalCount int64   `json:"total_count"`
}

type GetUserLikesResponseChan struct {
	Error *rpc.RpcError `json:"error"`
	Data  *LikedContent `json:"data"`
}
