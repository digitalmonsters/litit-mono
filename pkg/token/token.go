package token

import (
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetUserTokens(db *gorm.DB, userId int64) ([]database.Device, error) {
	var records []database.Device

	if err := db.Where("\"userId\" = ?", userId).Find(&records).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return records, nil
}

func CreateToken(db *gorm.DB, userId int64, req TokenCreateRequest) error {
	device := database.Device{
		UserId:    userId,
		DeviceId:  req.DeviceId,
		PushToken: req.PushToken,
		Platform:  req.Platform,
	}

	if err := db.Clauses(clause.OnConflict{UpdateAll: true, Columns: []clause.Column{
		{Name: "userId", Raw: true},
		{Name: "deviceId", Raw: true},
	}}).Create(&device).Error; err != nil {
		return err
	}

	return nil
}

func DeleteToken(db *gorm.DB, userId int64, deviceId string) error {
	var device database.Device

	if err := db.Where("device_id = ?", deviceId).Find(&device).Error; err != nil {
		return err
	}

	if device.Id.String() == uuid.Nil.String() || device.UserId != userId {
		return errors.WithStack(errors.New("device not found"))
	}

	if err := db.Delete(&device).Error; err != nil {
		return err
	}

	return nil
}
