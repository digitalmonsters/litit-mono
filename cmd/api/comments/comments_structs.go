package comments

import (
	"github.com/digitalmonsters/comments/pkg/comments"
	"gopkg.in/guregu/null.v4"
)

type updateCommentRequest struct {
	Comment string `json:"comment"`
}

type createCommentRequest struct {
	ParentId null.Int `json:"parent_id"`
	Comment  string   `json:"comment"`
}

type createCommentResponse struct {
	Id        int64  `json:"id"`
	Comment   string `json:"comment"`
	AuthorId  int64  `json:"author_id"`
	ContentId int64  `json:"content_id"`
}

type createCommentOnProfileResponse struct {
	Id        int64  `json:"id"`
	Comment   string `json:"comment"`
	AuthorId  int64  `json:"author_id"`
	ProfileId int64  `json:"profile_id"`
}

type successResponse struct {
	Success bool `json:"success"`
}

type frontendCommentResponse struct {
	comments.SimpleComment
	Author comments.Author `json:"author"`
}

type frontendCommentWithCursorResponse struct {
	comments.SimpleComment
	Author   comments.Author `json:"author"`
	Cursor   string          `json:"cursor"`
	ParentId null.Int        `json:"parent_id"`
}

type frontendCommentPaginationResponse struct {
	Comments []frontendCommentResponse `json:"comments"`
	Paging   comments.CursorPaging     `json:"paging"`
}
