package user_dislikes

import "github.com/digitalmonsters/go-common/rpc"

type GetAllUserDislikesRequest struct {
	UserId int64 `json:"user_id"`
}

type GetAllUserDislikesResponseChan struct {
	Error *rpc.RpcError `json:"error"`
	Data  []int64       `json:"data"`
}
