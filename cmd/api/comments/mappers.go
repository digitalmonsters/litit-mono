package comments

import (
	"github.com/digitalmonsters/comments/pkg/comments"
)

func commentToFrontendCommentResponse(comment comments.Comment) frontendCommentResponse {
	return frontendCommentResponse{
		SimpleComment: comment.SimpleComment,
		Author:        comment.Author,
	}
}

func commentsWithPagingToFrontendPaginationResponse(initialResult comments.GetCommentsByTypeWithResourceResponse) frontendCommentPaginationResponse {
	res := frontendCommentPaginationResponse{
		Paging: initialResult.Paging,
	}

	for _, c := range initialResult.Comments {
		res.Comments = append(res.Comments, commentToFrontendCommentResponse(c))
	}

	return res
}
