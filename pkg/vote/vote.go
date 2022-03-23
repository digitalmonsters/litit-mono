package vote

import (
	"github.com/digitalmonsters/comments/cmd/api/comments/notifiers/comment"
	"github.com/digitalmonsters/comments/cmd/api/vote/notifiers/vote"
	"github.com/digitalmonsters/comments/pkg/comments"
	"github.com/digitalmonsters/comments/pkg/database"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/pkg/errors"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func VoteComment(db *gorm.DB, commentId int64, voteUp null.Bool, currentUserId int64, commentNotifier *comment.Notifier,
	voteNotifier *vote.Notifier, apmTransaction *apm.Transaction,
	contentWrapper content.IContentWrapper) (*database.CommentVote, error) {
	tx := db.Begin()
	defer tx.Rollback()

	var commentToVote database.Comment

	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Take(&commentToVote, commentId).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	previousVoteValue := null.NewBool(false, false)
	var previousVote database.CommentVote

	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ? and comment_id = ?", currentUserId, commentId).
		Find(&previousVote).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	if previousVote.UserId > 0 { // if 0, then its new vote
		previousVoteValue = previousVote.VoteUp
	}

	previousVote.CommentId = commentId
	previousVote.UserId = currentUserId

	if !voteUp.Valid && !previousVote.VoteUp.Valid {
		return nil, nil
	}

	if voteUp.Valid && previousVote.VoteUp.Valid {
		if previousVote.VoteUp.ValueOrZero() == voteUp.ValueOrZero() { // nothing to do here, as status in db is already valid
			return nil, nil
		}
	}

	previousVote.VoteUp = voteUp

	if previousVote.VoteUp.Valid {
		if previousVote.VoteUp.ValueOrZero() {
			if err := tx.Model(&commentToVote).Update("num_upvotes", gorm.Expr("num_upvotes + 1")).Error; err != nil {
				return nil, errors.WithStack(err)
			}
		} else {
			if err := tx.Model(&commentToVote).Update("num_downvotes", gorm.Expr("num_downvotes + 1")).Error; err != nil {
				return nil, errors.WithStack(err)
			}
		}
	}

	if previousVoteValue.Valid {
		if previousVoteValue.Bool {
			if err := tx.Model(&commentToVote).Update("num_upvotes", gorm.Expr("num_upvotes - 1")).Error; err != nil {
				return nil, errors.WithStack(err)
			}
		} else {
			if err := tx.Model(&commentToVote).Update("num_downvotes", gorm.Expr("num_downvotes - 1")).Error; err != nil {
				return nil, errors.WithStack(err)
			}
		}
	}

	if err := tx.Exec("insert into comment_vote(user_id, comment_id, vote_up) values (?, ?, ?) on conflict (user_id, comment_id) do update set vote_up = ?", currentUserId, commentId, voteUp, voteUp).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	mapped := comments.MapDbCommentToComment(commentToVote)

	if commentNotifier != nil {
		var eventType eventsourcing.CommentChangeReason

		if !mapped.ContentId.IsZero() {
			eventType = eventsourcing.CommentChangeReasonContent

			extenders := []chan error{
				comments.ExtendWithContent(contentWrapper, apmTransaction, &mapped),
			}

			for _, e := range extenders {
				if err := <-e; err != nil {
					apm_helper.CaptureApmError(err, apmTransaction)
				}
			}
		} else {
			eventType = eventsourcing.CommentChangeReasonProfile
		}

		commentNotifier.Enqueue(commentToVote, mapped.Content, eventsourcing.ChangeEventTypeUpdated, eventType)

	}

	if voteNotifier != nil {
		if commentToVote.ContentId.Valid {
			voteNotifier.Enqueue(commentToVote.Id, currentUserId, voteUp, commentToVote.ParentId, commentToVote.AuthorId,
				commentToVote.Comment, commentToVote.ContentId.Int64, eventsourcing.ChangeEventTypeCreated, eventsourcing.CommentChangeReasonContent)
		} else if commentToVote.ProfileId.Valid {
			voteNotifier.Enqueue(commentToVote.Id, currentUserId, voteUp, commentToVote.ParentId, commentToVote.AuthorId,
				commentToVote.Comment, commentToVote.ProfileId.Int64, eventsourcing.ChangeEventTypeCreated, eventsourcing.CommentChangeReasonProfile)
		}
	}

	return &previousVote, nil
}
