package auth

import "github.com/digitalmonsters/go-common/rpc"

type AuthParseTokenRequest struct {
	Token            string `json:"token"`
	IgnoreExpiration bool   `json:"ignore_expiration"`
}

type AuthParseTokenResponseChan struct {
	Resp  AuthParseTokenResponse
	Error *rpc.RpcError
}

type AuthParseTokenResponse struct {
	UserId  int64 `json:"user_id"`
	Expired bool  `json:"expired"`
}
