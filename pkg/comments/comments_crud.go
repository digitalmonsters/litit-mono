package comments

import (
	"context"
	"github.com/digitalmonsters/comments/cmd/api/comments/notifiers/comment"
	"github.com/digitalmonsters/comments/cmd/api/comments/notifiers/content_comments_counter"
	"github.com/digitalmonsters/comments/cmd/api/comments/notifiers/user_comments_counter"
	"github.com/digitalmonsters/comments/pkg/database"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/pkg/errors"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func CreateComment(db *gorm.DB, resourceId int64, commentStr string, parentId null.Int, contentWrapper content.IContentWrapper,
	userWrapper user_go.IUserGoWrapper, ctx context.Context, currentUserId int64,
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
		ExtendWithContent(contentWrapper, ctx, &mappedParentComment),
	}

	for _, e := range extenders {
		if err := <-e; err != nil {
			apm_helper.LogError(err, ctx)
		}
	}

	blockedUserType, err := isBlocked(userWrapper, ctx, currentUserId, mappedParentComment.AuthorId)

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

	if err = updateUserStatsComments(tx, currentUserId, resourceId, true,
		userCommentsNotifier, contentCommentsNotifier, false); err != nil {
		return nil, err
	}

	if err = tx.Model(database.Content{}).Where("id = ?", resourceId).
		Update("comments_count", gorm.Expr("comments_count + 1")).Error; err != nil {
		return nil, err
	}

	if err = tx.Commit().Error; err != nil {
		return nil, err
	}

	mapped := MapDbCommentToComment(newComment)

	if commentNotifier != nil {
		extenders = []chan error{
			ExtendWithContent(contentWrapper, ctx, &mapped),
		}

		for _, e := range extenders {
			if err := <-e; err != nil {
				apm_helper.LogError(err, ctx)
			}
		}

		commentNotifier.Enqueue(newComment, eventsourcing.ChangeEventTypeCreated, eventsourcing.CommentChangeReasonContent)
		commentNotifier.Enqueue(parentComment, eventsourcing.ChangeEventTypeUpdated, eventsourcing.CommentChangeReasonContent)
	}

	return &mapped.SimpleComment, nil
}

func UpdateCommentById(db *gorm.DB, commentId int64, updatedComment string, currentUserId int64,
	contentWrapper content.IContentWrapper, commentNotifier *comment.Notifier,
	ctx context.Context) (*SimpleComment, error) {
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
		var eventType eventsourcing.CommentChangeReason

		if !mapped.ContentId.IsZero() {
			eventType = eventsourcing.CommentChangeReasonContent

			extenders := []chan error{
				ExtendWithContent(contentWrapper, ctx, &mapped),
			}

			for _, e := range extenders {
				if err := <-e; err != nil {
					apm_helper.LogError(err, ctx)
				}
			}
		} else {
			eventType = eventsourcing.CommentChangeReasonProfile
		}

		commentNotifier.Enqueue(modifiedComment, eventsourcing.ChangeEventTypeUpdated, eventType)
	}

	return &mapped.SimpleComment, nil
}

func DeleteComments(commentsToDelete []database.Comment, tx *gorm.DB, contentCommentsNotifier *content_comments_counter.Notifier,
	userCommentsNotifier *user_comments_counter.Notifier, commentNotifier *comment.Notifier) ([]func(), error) {
	parentIds := map[int64]int{}
	authors := map[int64]int{}
	contents := map[int64]int{}

	for _, c := range commentsToDelete {
		if err := tx.Exec("delete from comment where id = ?", c.Id).Error; err != nil {
			return nil, errors.WithStack(err)
		}

		if c.ParentId.Valid {
			parentIds[c.ParentId.ValueOrZero()] = parentIds[c.ParentId.ValueOrZero()] + 1
		}

		if c.AuthorId > 0 {
			authors[c.AuthorId] = authors[c.AuthorId] + 1
		}

		if c.ContentId.Valid {
			contents[c.ContentId.ValueOrZero()] = contents[c.ContentId.ValueOrZero()] + 1
		}
	}

	var callbacks []func()

	for parentId, count := range parentIds {
		if err := tx.Exec("update comment set num_replies = num_replies - ? where id = ?", count, parentId).
			Error; err != nil {
			return nil, err
		}
	}

	for commentAuthorId, count := range authors {
		var userStatsActionComments int64

		if err := tx.Raw("insert into user_stats_action(id, comments) values (?, 1) on conflict (id) "+
			"do update set comments = user_stats_action.comments - ? returning comments", commentAuthorId, count).
			Scan(&userStatsActionComments).Error; err != nil {
			return nil, err
		}

		callbacks = append(callbacks, func() {
			if userCommentsNotifier == nil {
				return
			}

			userCommentsNotifier.Enqueue(commentAuthorId, userStatsActionComments)
		})
	}

	for contentId, count := range contents {
		var userStatsContentComments int64

		if err := tx.Raw("insert into user_stats_content(id, comments) values (?, 1) on conflict (id) "+
			"do update set comments = user_stats_content.comments - ? returning comments", contentId, count).
			Scan(&userStatsContentComments).Error; err != nil {
			return nil, err
		}

		callbacks = append(callbacks, func() {
			if contentCommentsNotifier == nil {
				return
			}

			contentCommentsNotifier.Enqueue(contentId, userStatsContentComments)
		})
	}

	for _, c := range commentsToDelete {
		var eventType eventsourcing.CommentChangeReason

		if !c.ContentId.IsZero() {
			eventType = eventsourcing.CommentChangeReasonContent
		} else {
			eventType = eventsourcing.CommentChangeReasonProfile
		}

		callbacks = append(callbacks, func() {
			if commentNotifier == nil {
				return
			}

			commentNotifier.Enqueue(c, eventsourcing.ChangeEventTypeDeleted, eventType)
		})
	}

	return callbacks, nil
}

func DeleteCommentById(db *gorm.DB, commentId int64, currentUserId int64, contentWrapper content.IContentWrapper,
	apmTransaction *apm.Transaction, commentNotifier *comment.Notifier, contentCommentsNotifier *content_comments_counter.Notifier,
	userCommentsNotifier *user_comments_counter.Notifier) (*SimpleComment, error) {
	tx := db.Begin()
	defer tx.Rollback()

	var commentsToDelete []database.Comment // is raw list of all comments

	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? or parent_id = ?", commentId, commentId).
		Find(&commentsToDelete).Error; err != nil {
		return nil, err
	}

	if len(commentsToDelete) == 0 {
		return nil, errors.New("no comments to delete")
	}

	var mainComment database.Comment

	for _, c := range commentsToDelete {
		if c.Id == commentId {
			mainComment = c
		}
	}

	if mainComment.Id == 0 {
		return nil, errors.New("no main comment")
	}

	if mainComment.AuthorId != currentUserId {
		if mainComment.ProfileId.ValueOrZero() != currentUserId {
			contentId := mainComment.ContentId.ValueOrZero()

			if contentId == 0 {
				return nil, errors.New("delete operation not permitted")
			}

			resp := <-contentWrapper.GetInternal([]int64{contentId}, true,
				apmTransaction, false)

			if resp.Error != nil {
				return nil, resp.Error.ToError()
			}

			if v, ok := resp.Response[contentId]; !ok {
				return nil, errors.New("content not found")
			} else if v.AuthorId != currentUserId {
				return nil, errors.New("delete operation not permitted 2")
			}
		}
	}

	callbacks, err := DeleteComments(commentsToDelete, tx, contentCommentsNotifier,
		userCommentsNotifier, commentNotifier)

	if err != nil {
		return nil, err
	}

	if err = tx.Commit().Error; err != nil {
		return nil, errors.WithStack(err)
	}

	for _, c := range callbacks {
		c()
	}

	mapped := MapDbCommentToComment(mainComment).SimpleComment

	return &mapped, nil
}

func CreateCommentOnProfile(db *gorm.DB, resourceId int64, commentStr string, parentId null.Int,
	userGoWrapper user_go.IUserGoWrapper, ctx context.Context, currentUserId int64,
	commentNotifier *comment.Notifier, contentCommentsNotifier *content_comments_counter.Notifier,
	userCommentsNotifier *user_comments_counter.Notifier) (*SimpleComment, error) {
	var parentComment database.Comment

	if !parentId.IsZero() {
		if err := db.Take(&parentComment, parentId.ValueOrZero()).Error; err != nil {
			return nil, err
		}
	}

	mappedParentComment := mapDbCommentToCommentOnProfile(parentComment)

	blockedUserType, err := isBlocked(userGoWrapper, ctx, currentUserId, mappedParentComment.AuthorId)

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

	if err = updateUserStatsComments(tx, currentUserId, resourceId, false,
		userCommentsNotifier, contentCommentsNotifier, false); err != nil {
		return nil, err
	}

	if err = tx.Commit().Error; err != nil {
		return nil, err
	}

	mapped := mapDbCommentToCommentOnProfile(newComment)

	if commentNotifier != nil {
		commentNotifier.Enqueue(newComment, eventsourcing.ChangeEventTypeCreated, eventsourcing.CommentChangeReasonProfile)

		if !parentId.IsZero() {
			commentNotifier.Enqueue(parentComment, eventsourcing.ChangeEventTypeUpdated, eventsourcing.CommentChangeReasonProfile)
		}
	}

	return &mapped.SimpleComment, nil
}
