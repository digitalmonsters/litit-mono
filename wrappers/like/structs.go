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

//goland:noinspection GoNameStartsWithPackageName
type GetInternalLikedByUserResponseChan struct {
	Error *rpc.RpcError  `json:"error"`
	Data  map[int64]bool `json:"data"`
}

//goland:noinspection GoNameStartsWithPackageName
type GetInternalLikedByUserRequest struct {
	UserId     int64   `json:"user_id"`
	ContentIds []int64 `json:"content_ids"`
}

//goland:noinspection GoNameStartsWithPackageName
type GetInternalUserLikesResponseChan struct {
	Error           *rpc.RpcError `json:"error"`
	LikedContentIds []int64       `json:"liked_content_ids"`
}

type GetInternalUserLikesRequest struct {
	UserId int64 `json:"user_id"`
}

type getInternalUserLikesResponse struct {
	ContentIds []int64 `json:"content_ids"`
}
