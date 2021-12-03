package publicapi

import (
	"github.com/digitalmonsters/comments/pkg/database"
	"time"
)

func dbCommentToComment(comment database.Comment) Comment {
	return Comment{
		Id:           comment.,
		AuthorId:     0,
		NumReplies:   0,
		NumUpvotes:   0,
		NumDownvotes: 0,
		CreatedAt:    time.Time{},
		Author:       Author{},
	}
}