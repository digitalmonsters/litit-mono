package like

import (
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/rpc"
)

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

type GetInternalDislikedByUserResponseChan struct {
	Error *rpc.RpcError  `json:"error"`
	Data  map[int64]bool `json:"data"`
}

type SpotReaction struct {
	Like    bool
	Dislike bool
	Love    bool
}
type GetInternalSpotReactionsByUserResponseChan struct {
	Error *rpc.RpcError          `json:"error"`
	Data  map[int64]SpotReaction `json:"data"`
}

//goland:noinspection GoNameStartsWithPackageName
type GetInternalLikedByUserRequest struct {
	UserId     int64   `json:"user_id"`
	ContentIds []int64 `json:"content_ids"`
}

//goland:noinspection GoNameStartsWithPackageName
type GetInternalSpotReactionsByUserRequest struct {
	UserId     int64   `json:"user_id"`
	ContentIds []int64 `json:"content_ids"`
}

//goland:noinspection GoNameStartsWithPackageName
type GetInternalUserLikesResponseChan struct {
	Error           *rpc.RpcError `json:"error"`
	LikedContentIds []int64       `json:"liked_content_ids"`
}

type GetInternalUserLikesRequest struct {
	UserId    int64  `json:"user_id"`
	Size      int    `json:"size"`
	PageState string `json:"page_state"`
}

type getInternalUserLikesResponse struct {
	ContentIds []int64 `json:"content_ids"`
}

//goland:noinspection GoNameStartsWithPackageName
type GetInternalDislikedByUserRequest struct {
	UserId     int64   `json:"user_id"`
	ContentIds []int64 `json:"content_ids"`
}

type AddLikesRequest struct {
	LikeEvents []eventsourcing.LikeEvent `json:"like_events"`
}

type AddLikesResponse struct {
	Success bool `json:"success"`
}
