package onlineactivity

import (
	"fmt"

	"gorm.io/gorm"
)

type OnlineUserActivity struct {
	UserID   int64 `gorm:"primaryKey"`
	IsOnline bool
}

func TriggerUserOffline(db *gorm.DB, userId string) bool {
	result := db.Table("online_user_activity").Where("user_id = ?", userId).Update("is_online", false)
	if result.Error != nil {
		fmt.Printf("failed to update user online status: %v\n", result.Error)
		return false
	}
	return true
}

func TriggerUserOnline(db *gorm.DB, userId int64) bool {
	q := db.Table("online_user_activity").Where("user_id = ?", userId)
	if q.RowsAffected == 0 {
		result := db.Table("online_user_activity").Create(&OnlineUserActivity{UserID: userId, IsOnline: true})
		if result.Error != nil {
			fmt.Printf("failed to create online user activity: %v\n", result.Error)
			return false
		}
		return true
	}
	result := db.Table("online_user_activity").Where("user_id = ?", userId).Update("is_online", true)
	if result.Error != nil {
		fmt.Printf("failed to update user online status: %v\n", result.Error)
		return false
	}
	return true
}
