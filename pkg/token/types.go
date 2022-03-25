package token

import "github.com/digitalmonsters/go-common/common"

type TokenCreateRequest struct {
	DeviceId  string            `json:"device_id"`
	PushToken string            `json:"push_token"`
	Platform  common.DeviceType `json:"platform"`
}
