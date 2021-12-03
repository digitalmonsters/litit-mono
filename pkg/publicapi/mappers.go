package publicapi

import (
	"github.com/digitalmonsters/comments/pkg/database"
)

func mapDbCommentToComment(comment database.Comment) Comment {
	return Comment{
		Id:           comment.Id,
		AuthorId:     comment.AuthorId,
		NumReplies:   comment.NumReplies,
		NumUpvotes:   comment.NumUpvotes,
		NumDownvotes: comment.NumDownvotes,
		CreatedAt:    comment.CreatedAt,
	}
}
