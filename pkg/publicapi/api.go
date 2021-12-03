package publicapi

import (
	"github.com/digitalmonsters/comments/pkg/database"
	"gorm.io/gorm"
)

func GetCommendById(commentId int64, db *gorm.DB, currentUserId int64) (interface{}, error) {
	var comment database.Comment

	if err := db.Find(&comment).Take(&comment, commentId).Error; err != nil {
		return nil, err
	}

	if currentUserId > 0 {
		if err :=
	}
}
