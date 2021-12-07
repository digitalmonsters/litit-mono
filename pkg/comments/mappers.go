package comments

import (
	"github.com/digitalmonsters/comments/pkg/database"
)

func mapDbCommentToComment(comment database.Comment) Comment {
	return Comment{
		SimpleComment: SimpleComment{
			Id:           comment.Id,
			AuthorId:     comment.AuthorId,
			NumReplies:   comment.NumReplies,
			NumUpvotes:   comment.NumUpvotes,
			NumDownvotes: comment.NumDownvotes,
			CreatedAt:    comment.CreatedAt,
			ContentId:    comment.ContentId,
			Comment:      comment.Comment,
		},
	}
}

func mapDbCommentToCommentOnProfile(comment database.Comment) CommentOnProfile {
	return CommentOnProfile{
		SimpleComment: SimpleComment{
			Id:           comment.Id,
			AuthorId:     comment.AuthorId,
			NumReplies:   comment.NumReplies,
			NumUpvotes:   comment.NumUpvotes,
			NumDownvotes: comment.NumDownvotes,
			CreatedAt:    comment.CreatedAt,
			ContentId:    comment.ContentId,
			Comment:      comment.Comment,
		},
	}
}
