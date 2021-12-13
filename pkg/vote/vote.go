package vote

import (
	"github.com/digitalmonsters/comments/pkg/database"
	"github.com/pkg/errors"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func VoteComment(db *gorm.DB, commentId int64, voteUp null.Bool, currentUserId int64) (*database.CommentVote, error) {
	tx := db.Begin()
	defer tx.Rollback()

	var comment database.Comment

	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Take(&comment, commentId).Error; err != nil {
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
			if err := tx.Model(&comment).Update("num_upvotes", gorm.Expr("num_upvotes + 1")).Error; err != nil {
				return nil, errors.WithStack(err)
			}
		} else {
			if err := tx.Model(&comment).Update("num_downvotes", gorm.Expr("num_downvotes + 1")).Error; err != nil {
				return nil, errors.WithStack(err)
			}
		}
	}

	if previousVoteValue.Valid {
		if previousVoteValue.Bool {
			if err := tx.Model(&comment).Update("num_upvotes", gorm.Expr("num_upvotes - 1")).Error; err != nil {
				return nil, errors.WithStack(err)
			}
		} else {
			if err := tx.Model(&comment).Update("num_downvotes", gorm.Expr("num_downvotes - 1")).Error; err != nil {
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

	return &previousVote, nil
}
