package publicapi

import (
	"errors"
	"fmt"
	"github.com/digitalmonsters/comments/pkg/database"
	"github.com/digitalmonsters/go-common/wrappers/user"
	"github.com/thoas/go-funk"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
)

func extendWithLikedByMe(db *gorm.DB, currentUserId int64, comments ...*Comment) chan error {
	ch := make(chan error)

	if currentUserId == 0 || len(comments) == 0 {
		close(ch)

		return ch
	}

	go func() {
		defer func() {
			close(ch)
		}()

		var commentIds []int64

		for _, comment := range comments {
			commentIds = append(commentIds, comment.Id)
		}

		var foundVotes []struct {
			CommentId int64
			VoteUp    bool
		}

		if err := db.Model(database.CommentVote{}).Where("user_id = ? and comment_id in ?", currentUserId, commentIds).
			Select("comment_id").Find(&foundVotes).Error; err != nil { // todo check mapping
			ch <- err
			return
		}

		voteMap := map[int64]bool{}

		for _, f := range foundVotes {
			voteMap[f.CommentId] = f.VoteUp
		}

		if len(foundVotes) > 0 {
			for _, comment := range comments {
				if v, ok := voteMap[comment.Id]; ok {
					comment.MyVoteUp = null.BoolFrom(v)
				}
			}
		}
	}()

	return ch
}
func extendWithAuthor(userInfoWrapper user.IUserWrapper, apmTransaction *apm.Transaction, comments ...*Comment) chan error {
	ch := make(chan error)

	if len(comments) == 0 {
		close(ch)

		return ch
	}

	go func() {
		defer func() {
			close(ch)
		}()

		var authors []int64

		for _, comment := range comments {
			if !funk.ContainsInt64(authors, comment.AuthorId) {
				authors = append(authors, comment.AuthorId)
			}
		}

		if len(authors) > 0 {
			responseData := <-userInfoWrapper.GetUsers(authors, apmTransaction, false)

			if responseData.Error != nil {
				ch <- errors.New(fmt.Sprintf("invalid response from user service [%v]", responseData.Error.Message))
				return
			}

			for _, comment := range comments {
				if v, ok := responseData.Items[comment.AuthorId]; ok {
					comment.Author = Author{
						Id:        v.UserId,
						Username:  v.Username,
						Avatar:    v.Avatar,
						Firstname: v.Firstname,
						Lastname:  v.Lastname,
					}
				}
			}
		}
	}()

	return ch
}
