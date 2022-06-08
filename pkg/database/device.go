package database

import (
	"github.com/digitalmonsters/go-common/common"
	"github.com/google/uuid"
)

type Device struct {
	Id        uuid.UUID         `gorm:"primaryKey;autoIncrement"`
	UserId    int64             `gorm:"column:userId"`
	DeviceId  string            `gorm:"column:deviceId"`
	PushToken string            `gorm:"column:pushToken"`
	Platform  common.DeviceType `json:"platform"`
}

func (Device) TableName() string {
	return "devices"
}
