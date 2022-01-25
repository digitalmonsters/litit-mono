package auth_go

import (
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/rpc"
)

type CheckAdminPermissionsRequest struct {
	UserId      int64              `json:"jwt"`
	Method      string             `json:"method"`
	AccessLevel common.AccessLevel `json:"access_level"`
}

type CheckLegacyAdminResponseChan struct {
	Error *rpc.RpcError
	Resp  CheckLegacyAdminResponse
}

type CheckLegacyAdminResponse struct {
	IsAdmin      bool `json:"is_admin"`
	IsSuperAdmin bool `json:"is_super_admin"`
}

type CheckLegacyAdminRequest struct {
	UserId int64 `json:"user_id"`
}

type CheckAdminPermissionsResponseChan struct {
	Resp  CheckAdminPermissionsResponse
	Error *rpc.RpcError
}

type CheckAdminPermissionsResponse struct {
	UserId    int64 `json:"user_id"`
	HasAccess bool  `json:"has_access"`
}