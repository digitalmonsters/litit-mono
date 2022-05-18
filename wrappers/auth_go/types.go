package auth_go

import (
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/rpc"
	"gopkg.in/guregu/null.v4"
)

type IsGuestRequest struct {
	UserId int64 `json:"user_id"`
}

type IsGuestResponse struct {
	IsGuest bool `json:"is_guest"`
}

type AddUserResponseChan struct {
	Error *rpc.RpcError
	Item  eventsourcing.UserEvent `json:"item"`
}

type CheckAdminPermissionsRequest struct {
	UserId int64  `json:"user_id"`
	Object string `json:"object"`
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

type GetAdminIdsFilterByEmailRequest struct {
	AdminIds    []int64 `json:"admin_ids"`
	SearchQuery string  `json:"search_query"`
}

type GetAdminIdsFilterByEmailResponseChan struct {
	Error    *rpc.RpcError
	AdminIds []int64 `json:"admin_ids"`
}

type GetAdminsInfoByIdRequest struct {
	AdminIds []int64 `json:"admin_ids"`
}

type GetAdminsInfoByIdResponseChan struct {
	Error *rpc.RpcError
	Items map[int64]AdminGeneralInfo `json:"items"`
}

type AdminGeneralInfo struct {
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Name        string `json:"name"`
	Country     string `json:"country"`
}

type GetUsersRegistrationTypeRequest struct {
	UserIds []int64 `json:"user_ids"`
}

type SocialProviderType int

const (
	SocialProviderTypeGoogle   = SocialProviderType(1)
	SocialProviderTypeApple    = SocialProviderType(2)
	SocialProviderTypeFacebook = SocialProviderType(3)
	SocialProviderTypeEmail    = SocialProviderType(4)
)

type InternalGetUsersForValidatorFromCacheRequest struct {
	UserIds []int64 `json:"user_ids"`
}

type UserForValidator struct {
	Id         int64     `json:"id"`
	Deleted    bool      `json:"deleted"`
	BannedTill null.Time `json:"banned_till"`
	Guest      bool      `json:"guest"`
	Verified   bool      `json:"verified"`
}
