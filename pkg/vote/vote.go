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

	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Find(&previousVote, commentId, currentUserId).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	if previousVote.UserId > 0 { // if 0, then its new vote
		previousVoteValue = null.BoolFrom(previousVote.VoteUp)
	}

	previousVote.CommentId = commentId
	previousVote.UserId = currentUserId

	if previousVote.VoteUp == voteUp.ValueOrZero() { // nothing to do here, as status in db is already valid
		return nil, nil
	}

	previousVote.VoteUp = voteUp.ValueOrZero()

	if previousVote.VoteUp {
		if err := tx.Model(&comment).Update("num_upvotes", gorm.Expr("num_upvotes + 1")).Error; err != nil {
			return nil, errors.WithStack(err)
		}
	} else {
		if err := tx.Model(&comment).Update("num_downvotes", gorm.Expr("num_upvotes + 1")).Error; err != nil {
			return nil, errors.WithStack(err)
		}
	}

	if previousVoteValue.Valid {
		if previousVoteValue.Bool {
			if err := tx.Model(&comment).Update("num_upvotes", gorm.Expr("num_upvotes - 1")).Error; err != nil {
				return nil, errors.WithStack(err)
			}
		} else {
			if err := tx.Model(&comment).Update("num_downvotes", gorm.Expr("num_upvotes - 1")).Error; err != nil {
				return nil, errors.WithStack(err)
			}
		}
	}

	if err := tx.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Save(previousVote).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &previousVote, nil
}
