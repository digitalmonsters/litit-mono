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

func mapDbCommentForDeleteToCommentForDelete(comment database.CommentForDelete) CommentForDelete {
	return CommentForDelete{
		Id:         comment.Id,
		AuthorId:   comment.AuthorId,
		NumReplies: comment.NumReplies,
		ContentId:  comment.ContentId,
		ParentId:   comment.ParentId,
		Content:    ContentWithAuthorId{},
	}
}

func mapDbCommentForSendToCommentForSend(comment database.Comment) CommentForSend {
	return CommentForSend{
		Id:         comment.Id,
		AuthorId:   comment.AuthorId,
		NumReplies: comment.NumReplies,
		ContentId:  comment.ContentId,
		ParentId:   comment.ParentId,
		Content:    ContentWithAuthorId{},
	}
}
