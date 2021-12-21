package follow

import "github.com/digitalmonsters/go-common/rpc"

type GetUserFollowingRelationBulkResponseChan struct {
	Error *rpc.RpcError          `json:"error"`
	Data  map[int64]RelationData `json:"data"`
}

type GetUserFollowingRelationBulkRequest struct {
	UserId         int64   `json:"user_id"`
	RequestUserIds []int64 `json:"request_user_ids"`
}

type GetUserFollowingRelationRequest struct {
	UserId        int64 `json:"user_id"`
	RequestUserId int64 `json:"request_user_id"`
}

type GetUserFollowingRelationResponseChan struct {
	Error       *rpc.RpcError `json:"error"`
	IsFollower  bool          `json:"is_follower"`
	IsFollowing bool          `json:"is_following"`
}

type RelationData struct {
	IsFollower  bool `json:"is_follower"`
	IsFollowing bool `json:"is_following"`
}

type UserFollowingRelationResponse struct {
	Data map[int64]RelationData `json:"data"`
}

type GetUserFollowersRequest struct {
	UserId    int64  `json:"user_id"`
	PageState string `json:"page_state"`
	Limit     int    `json:"limit"`
}

type GetUserFollowersResponse struct {
	FollowerIds []int64 `json:"follower_ids"`
	PageState   string  `json:"page_state"`
}

type GetUserFollowersResponseChan struct {
	Error       *rpc.RpcError `json:"error"`
	FollowerIds []int64       `json:"follower_ids"`
	PageState   string        `json:"page_state"`
}

type GetFollowersCountRequest struct {
	UserIds []int64 `json:"user_ids"`
}

type GetFollowersCountResponseChan struct {
	Error *rpc.RpcError   `json:"error"`
	Data  map[int64]int64 `json:"data"`
}

type SimpleFollowing struct {
	FollowerUserId int64
	Follow         bool
	UserId         int64
}
