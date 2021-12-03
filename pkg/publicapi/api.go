package publicapi

import (
	"github.com/digitalmonsters/comments/pkg/database"
	"github.com/digitalmonsters/go-common/wrappers/user"
	"go.elastic.co/apm"
	"gorm.io/gorm"
)

func GetCommendById(db *gorm.DB, commentId int64, currentUserId int64, userWrapper user.IUserWrapper,
	apmTransaction *apm.Transaction) (interface{}, error) {
	var comment *database.Comment

	if err := db.Find(&comment).Take(&comment, commentId).Error; err != nil {
		return nil, err
	}

	if currentUserId > 0 {
		if err := db.Model(database.CommentVote{}).Where("user_id = ? and comment_id = ?",
			currentUserId, commentId)
	}

	extendWithAuthor(userWrapper, apmTransaction, comment)
}