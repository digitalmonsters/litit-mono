package user_block

import (
	"github.com/digitalmonsters/go-common/rpc"
)

type GetBlockListResponseChan struct {
	Error *rpc.RpcError
	Data  map[int64][]int64 `json:"data"`
}

type GetBlockListRequest struct {
	UserIds []int64 `json:"user_ids"`
}

type GetUserBlockResponseChan struct {
	Error *rpc.RpcError
	Data  UserBlockData `json:"data"`
}

type UserBlockData struct {
	Type      *BlockedUserType `json:"type"`
	IsBlocked bool             `json:"is_blocked"`
}

type BlockedUserType string

const (
	BlockedUser   BlockedUserType = "BLOCKED_BY_USER"
	BlockedByUser BlockedUserType = "BLOCKED_TO_USER"
)

type GetUserBlockRequest struct {
	BlockBy   int64 `json:"block_by"`
	BlockedTo int64 `json:"blocked_to"`
}
