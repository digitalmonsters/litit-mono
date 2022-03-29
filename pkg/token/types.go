package token

import (
	"github.com/digitalmonsters/go-common/common"
	"github.com/google/uuid"
	"gopkg.in/guregu/null.v4"
	"time"
)

type TokenCreateRequest struct {
	DeviceId  string            `json:"device_id"`
	PushToken string            `json:"push_token"`
	Platform  common.DeviceType `json:"platform"`
}

type TokenCreateResponse struct {
	Id        uuid.UUID         `json:"id"`
	UserId    int64             `json:"userId"`
	DeviceId  string            `json:"deviceId"`
	PushToken string            `json:"pushToken"`
	Platform  common.DeviceType `json:"platform"`
	CreatedAt time.Time         `json:"createdAt"`
	UpdatedAt null.Time         `json:"updatedAt"`
	DeletedAt null.Time         `json:"deletedAt"`
}
