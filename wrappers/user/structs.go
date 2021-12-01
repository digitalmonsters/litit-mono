package user

import (
	"github.com/digitalmonsters/go-common/rpc"
	"gopkg.in/guregu/null.v4"
)

type GetUsersResponseChan struct {
	Error *rpc.RpcError
	Items map[int64]UserRecord `json:"items"`
}

//goland:noinspection GoNameStartsWithPackageName
type UserRecord struct {
	UserId                     int64       `json:"user_id"`
	Avatar                     null.String `json:"avatar"`
	Username                   string      `json:"username"`
	Firstname                  string      `json:"firstname"`
	Lastname                   string      `json:"lastname"`
	Verified                   bool        `json:"verified"`
	EnableAgeRestrictedContent bool        `json:"enable_age_restricted_content"`
}

type GetUsersRequest struct {
	UserIds []int64 `json:"user_ids"`
}
