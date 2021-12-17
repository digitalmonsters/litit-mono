package follow

import "github.com/digitalmonsters/go-common/rpc"

//goland:noinspection GoNameStartsWithPackageName
type FollowContentUserByContentIdsResponseChan struct {
	Error *rpc.RpcError  `json:"error"`
	Data  map[int64]bool `json:"follow_statuses"`
}

//goland:noinspection GoNameStartsWithPackageName
type FollowContentUserByContentIdsRequest struct {
	UserId     int64   `json:"user_id"`
	ContentIds []int64 `json:"content_ids"`
}
