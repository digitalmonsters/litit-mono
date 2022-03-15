package token

import (
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func GetUserTokens(db *gorm.DB, userId int64) ([]database.Device, error) {
	var records []database.Device

	if err := db.Where("\"userId\" = ?", userId).Find(&records).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return records, nil
}
