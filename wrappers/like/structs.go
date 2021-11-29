package like

import "github.com/digitalmonsters/go-common/rpc"

type LastLikedByUserResponseChan struct {
	Error *rpc.RpcError          `json:"error"`
	Items map[int64][]LikeRecord `json:"items"`
}

//goland:noinspection GoNameStartsWithPackageName
type LikeRecord struct {
	ContentId int64 `json:"content_id"`
}

type GetLatestLikedByUserRequest struct {
	LimitPerUser int     `json:"limit_per_user"`
	UserIds      []int64 `json:"user_ids"`
}
