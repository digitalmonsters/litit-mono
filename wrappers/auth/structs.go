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
	IsGuest bool  `json:"is_guest"`
	Expired bool  `json:"expired"`
}

type GenerateTokenResponseChan struct {
	Resp  GenerateTokenResponse
	Error *rpc.RpcError
}

type GenerateTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type ParseAdminJwtTokenRequest struct {
	Jwt string `json:"jwt"`
}

type ParseAdminJwtTokenResponse struct {
	UserId int64 `json:"user_id"`
}

type ParseAdminJwtTokenResponseChan struct {
	Error *rpc.RpcError
	Resp  ParseAdminJwtTokenResponse
}
