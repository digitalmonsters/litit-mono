package user_block

import (
	"github.com/digitalmonsters/go-common/rpc"
)

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
	BlockedUser   BlockedUserType = "BLOCKED USER"
	BlockedByUser BlockedUserType = "YOUR PROFILE IS BLOCKED BY USER"
)

type GetUserBlockRequest struct {
	BlockBy   int64 `json:"block_by"`
	BlockedTo int64 `json:"blocked_to"`
}
