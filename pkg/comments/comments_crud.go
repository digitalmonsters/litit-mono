package comments

import (
	"github.com/digitalmonsters/comments/pkg/database"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/pkg/errors"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func CreateComment(db *gorm.DB, resourceId int64, commentStr string, parentId null.Int, contentWrapper content.IContentWrapper,
	apmTransaction *apm.Transaction, currentUserId int64) (*SimpleComment, error) {
	var parentComment database.Comment

	if !parentId.IsZero() {
		if err := db.Take(&parentComment, parentId.ValueOrZero()).Error; err != nil {
			return nil, err
		}
	}

	mappedComment := mapDbCommentToComment(parentComment)

	extenders := []chan error{
		extendWithContentForSend(contentWrapper, apmTransaction, &mappedComment),
	}

	for _, e := range extenders {
		if err := <-e; err != nil {
			apm_helper.CaptureApmError(err, apmTransaction)
		}
	}

	blockedUserType, err := isBlocked(currentUserId, mappedComment.AuthorId)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	if blockedUserType != nil {
		return nil, errors.WithStack(errors.New(string(*blockedUserType)))
	}

	tx := db.Begin()
	defer tx.Rollback()
	var comment database.Comment

	comment.ContentId = resourceId
	comment.Comment = commentStr
	comment.AuthorId = currentUserId
	comment.ParentId = parentId

	if err = tx.Omit("created_at").Create(&comment).Error; err != nil {
		return nil, err
	}

	if !parentId.IsZero() {
		if err := tx.Model(&parentComment).Update("num_replies", parentComment.NumReplies+1).Error; err != nil {
			return nil, err
		}
	}

	if err = updateUserStatsComments(currentUserId); err != nil {
		return nil, err
	}

	// TODO: send notify 'comment'

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	mapped := mapDbCommentToComment(comment)

	return &mapped.SimpleComment, nil
}

func UpdateCommentById(db *gorm.DB, commentId int64, updatedComment string, currentUserId int64) (*SimpleComment, error) {
	var comment database.Comment

	tx := db.Begin()
	defer tx.Rollback()

	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("author_id = ?", currentUserId).
		Take(&comment, commentId).Error; err != nil {
		return nil, err
	}

	if err := tx.Model(&comment).Update("comment", updatedComment).Error; err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.WithStack(err)
	}

	mapped := mapDbCommentToComment(comment)

	return &mapped.SimpleComment, nil
}

func DeleteCommentById(db *gorm.DB, commentId int64, currentUserId int64, contentWrapper content.IContentWrapper,
	apmTransaction *apm.Transaction) (*SimpleComment, error) {
	tx := db.Begin()
	defer tx.Rollback()

	var comment database.Comment

	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Take(&comment, commentId).Error; err != nil {
		return nil, err
	}

	mappedComment := mapDbCommentToComment(comment)

	extenders := []chan error{
		extendWithContentId(contentWrapper, apmTransaction, &mappedComment),
	}

	for _, e := range extenders {
		if err := <-e; err != nil {
			apm_helper.CaptureApmError(err, apmTransaction)
		}
	}

	if mappedComment.AuthorId != currentUserId && mappedComment.Content.AuthorId != currentUserId {
		return nil, errors.WithStack(errors.New("not allowed"))
	}

	if comment.ParentId.ValueOrZero() > 0 {
		if err := tx.Model(database.Comment{}).Where("id = ?", comment.ParentId.ValueOrZero()).
			Update("num_replies", gorm.Expr("num_replies - 1")).Error; err != nil {
			return nil, err
		}
	}

	if err := tx.Delete(&comment, commentId).Error; err != nil {
		return nil, err
	}

	return &mappedComment.SimpleComment, tx.Commit().Error
}
