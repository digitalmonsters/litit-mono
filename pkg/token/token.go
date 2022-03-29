package token

import (
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"time"
)

func GetUserTokens(db *gorm.DB, userId int64) ([]database.Device, error) {
	var records []database.Device

	if err := db.Where("\"userId\" = ?", userId).Find(&records).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return records, nil
}

func CreateToken(db *gorm.DB, userId int64, req TokenCreateRequest) (*TokenCreateResponse, error) {
	device := database.Device{
		UserId:    userId,
		DeviceId:  req.DeviceId,
		PushToken: req.PushToken,
		Platform:  req.Platform,
	}

	if err := db.Create(&device).Error; err != nil {
		return nil, err
	}

	return &TokenCreateResponse{
		Id:        device.Id,
		UserId:    device.UserId,
		DeviceId:  device.DeviceId,
		PushToken: device.PushToken,
		Platform:  device.Platform,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func DeleteToken(db *gorm.DB, userId int64, deviceId string) error {
	if err := db.Where("delete from \"Devices\" where \"deviceId\" = ? and \"userId\" = ?", deviceId, userId).Error; err != nil {
		return err
	}

	return nil
}
