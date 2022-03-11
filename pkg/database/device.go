package database

import (
	"github.com/digitalmonsters/go-common/wrappers/notification_gateway"
	"github.com/google/uuid"
)

type Device struct {
	Id        uuid.UUID
	UserId    int64                           `gorm:"column:userId"`
	DeviceId  string                          `gorm:"column:deviceId"`
	PushToken string                          `gorm:"column:pushToken"`
	Platform  notification_gateway.DeviceType `json:"platform"`
}

func (Device) TableName() string {
	return "Devices"
}
