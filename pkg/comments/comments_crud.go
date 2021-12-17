package comments

import (
	"github.com/digitalmonsters/comments/cmd/notifiers/comment"
	"github.com/digitalmonsters/comments/cmd/notifiers/content_comments_counter"
	"github.com/digitalmonsters/comments/cmd/notifiers/user_comments_counter"
	"github.com/digitalmonsters/comments/pkg/database"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/digitalmonsters/go-common/wrappers/user_block"
	"github.com/pkg/errors"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func CreateComment(db *gorm.DB, resourceId int64, commentStr string, parentId null.Int, contentWrapper content.IContentWrapper,
	userBlockWrapper user_block.IUserBlockWrapper, apmTransaction *apm.Transaction, currentUserId int64,
	commentNotifier *comment.Notifier, contentCommentsNotifier *content_comments_counter.Notifier,
	userCommentsNotifier *user_comments_counter.Notifier) (*SimpleComment, error) {
	var parentComment database.Comment

	tx := db.Begin()
	defer tx.Rollback()

	if !parentId.IsZero() {
		if err := tx.Take(&parentComment, parentId.ValueOrZero()).Error; err != nil {
			return nil, err
		}
	}

	mappedParentComment := MapDbCommentToComment(parentComment)

	extenders := []chan error{
		ExtendWithContent(contentWrapper, apmTransaction, &mappedParentComment),
	}

	for _, e := range extenders {
		if err := <-e; err != nil {
			apm_helper.CaptureApmError(err, apmTransaction)
		}
	}

	blockedUserType, err := isBlocked(userBlockWrapper, apmTransaction, currentUserId, mappedParentComment.AuthorId)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	if blockedUserType != nil {
		return nil, errors.WithStack(errors.New(string(*blockedUserType)))
	}

	var newComment database.Comment

	newComment.ContentId = null.IntFrom(resourceId)
	newComment.Comment = commentStr
	newComment.AuthorId = currentUserId
	newComment.ParentId = parentId
	newComment.Active = true

	if err = tx.Omit("created_at").Create(&newComment).Error; err != nil {
		return nil, err
	}

	if !parentId.IsZero() {
		if err := tx.Model(&parentComment).Update("num_replies", parentComment.NumReplies+1).Error; err != nil {
			return nil, err
		}
	}

	if err = updateUserStatsComments(tx, currentUserId, resourceId, userCommentsNotifier, contentCommentsNotifier); err != nil {
		return nil, err
	}

	if err = updateContentCommentsCounter(tx, newComment.ContentId.ValueOrZero(), true); err != nil {
		return nil, err
	}

	if err = tx.Commit().Error; err != nil {
		return nil, err
	}

	mapped := MapDbCommentToComment(newComment)

	if commentNotifier != nil {
		extenders = []chan error{
			ExtendWithContent(contentWrapper, apmTransaction, &mapped),
		}

		for _, e := range extenders {
			if err := <-e; err != nil {
				apm_helper.CaptureApmError(err, apmTransaction)
			}
		}

		commentNotifier.Enqueue(newComment, mapped.Content, comment.ContentResourceTypeCreate)
		commentNotifier.Enqueue(parentComment, mappedParentComment.Content, comment.ContentResourceTypeUpdate)
	}

	return &mapped.SimpleComment, nil
}

func UpdateCommentById(db *gorm.DB, commentId int64, updatedComment string, currentUserId int64,
	contentWrapper content.IContentWrapper, commentNotifier *comment.Notifier,
	apmTransaction *apm.Transaction) (*SimpleComment, error) {
	var modifiedComment database.Comment

	tx := db.Begin()
	defer tx.Rollback()

	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("author_id = ?", currentUserId).
		Take(&modifiedComment, commentId).Error; err != nil {
		return nil, err
	}

	if err := tx.Model(&modifiedComment).Update("comment", updatedComment).Error; err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.WithStack(err)
	}

	mapped := MapDbCommentToComment(modifiedComment)

	if commentNotifier != nil {
		var eventType comment.EventType

		if !mapped.ContentId.IsZero() {
			eventType = comment.ContentResourceTypeUpdate

			extenders := []chan error{
				ExtendWithContent(contentWrapper, apmTransaction, &mapped),
			}

			for _, e := range extenders {
				if err := <-e; err != nil {
					apm_helper.CaptureApmError(err, apmTransaction)
				}
			}
		} else {
			eventType = comment.ProfileResourceTypeUpdate
		}

		commentNotifier.Enqueue(modifiedComment, mapped.Content, eventType)
	}

	return &mapped.SimpleComment, nil
}

func DeleteCommentById(db *gorm.DB, commentId int64, currentUserId int64, contentWrapper content.IContentWrapper,
	apmTransaction *apm.Transaction, commentNotifier *comment.Notifier) (*SimpleComment, error) {
	tx := db.Begin()
	defer tx.Rollback()

	var commentToDelete database.Comment

	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Take(&commentToDelete, commentId).Error; err != nil {
		return nil, err
	}

	mappedComment := MapDbCommentToComment(commentToDelete)

	extenders := []chan error{
		ExtendWithContent(contentWrapper, apmTransaction, &mappedComment),
	}

	for _, e := range extenders {
		if err := <-e; err != nil {
			apm_helper.CaptureApmError(err, apmTransaction)
		}
	}

	if mappedComment.AuthorId != currentUserId && mappedComment.Content.AuthorId != currentUserId {
		return nil, errors.WithStack(errors.New("not allowed"))
	}

	var parentComment database.Comment

	if commentToDelete.ParentId.ValueOrZero() > 0 {
		if err := tx.Model(&parentComment).Where("id = ?", commentToDelete.ParentId.ValueOrZero()).
			Update("num_replies", gorm.Expr("num_replies - 1")).Scan(&parentComment).Error; err != nil {
			return nil, err
		}
	}

	if err := tx.Delete(&commentToDelete, commentId).Error; err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	if err := updateContentCommentsCounter(db, commentToDelete.ContentId.ValueOrZero(), false); err != nil {
		return nil, err
	}

	if commentNotifier != nil {
		var eventType comment.EventType

		if !mappedComment.ContentId.IsZero() {
			eventType = comment.ContentResourceTypeDelete
		} else {
			eventType = comment.ProfileResourceTypeDelete
		}

		commentNotifier.Enqueue(commentToDelete, mappedComment.Content, eventType)

		if !parentComment.ParentId.IsZero() {
			mappedParentComment := MapDbCommentToComment(parentComment)

			if eventType == comment.ContentResourceTypeDelete {
				extenders = []chan error{
					ExtendWithContent(contentWrapper, apmTransaction, &mappedParentComment),
				}

				for _, e := range extenders {
					if err := <-e; err != nil {
						apm_helper.CaptureApmError(err, apmTransaction)
					}
				}
			}

			commentNotifier.Enqueue(parentComment, mappedParentComment.Content, eventType)
		}
	}

	return &mappedComment.SimpleComment, nil
}

func CreateCommentOnProfile(db *gorm.DB, resourceId int64, commentStr string, parentId null.Int, contentWrapper content.IContentWrapper,
	userBlockWrapper user_block.IUserBlockWrapper, apmTransaction *apm.Transaction, currentUserId int64,
	commentNotifier *comment.Notifier, contentCommentsNotifier *content_comments_counter.Notifier,
	userCommentsNotifier *user_comments_counter.Notifier) (*SimpleComment, error) {
	var parentComment database.Comment

	if !parentId.IsZero() {
		if err := db.Take(&parentComment, parentId.ValueOrZero()).Error; err != nil {
			return nil, err
		}
	}

	mappedParentComment := mapDbCommentToCommentOnProfile(parentComment)

	blockedUserType, err := isBlocked(userBlockWrapper, apmTransaction, currentUserId, mappedParentComment.AuthorId)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	if blockedUserType != nil {
		return nil, errors.WithStack(errors.New(string(*blockedUserType)))
	}

	tx := db.Begin()
	defer tx.Rollback()
	var newComment database.Comment

	newComment.ProfileId = null.IntFrom(resourceId)
	newComment.Comment = commentStr
	newComment.AuthorId = currentUserId
	newComment.ParentId = parentId
	newComment.Active = true

	if err = tx.Omit("created_at").Create(&newComment).Error; err != nil {
		return nil, err
	}

	if !parentId.IsZero() {
		if err := tx.Model(&parentComment).Update("num_replies", parentComment.NumReplies+1).Error; err != nil {
			return nil, err
		}
	}

	if err = updateUserStatsComments(tx, currentUserId, resourceId, userCommentsNotifier, contentCommentsNotifier); err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	mapped := mapDbCommentToCommentOnProfile(newComment)

	if commentNotifier != nil {
		commentNotifier.Enqueue(newComment, content.SimpleContent{}, comment.ProfileResourceTypeCreate)

		if !parentId.IsZero() {
			commentNotifier.Enqueue(parentComment, content.SimpleContent{}, comment.ProfileResourceTypeUpdate)
		}
	}

	return &mapped.SimpleComment, nil
}
