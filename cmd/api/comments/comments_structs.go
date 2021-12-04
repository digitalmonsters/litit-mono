package comments

import (
	"github.com/digitalmonsters/comments/pkg/publicapi"
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

type successResponse struct {
	Success bool `json:"success"`
}

type frontendCommentResponse struct {
	publicapi.SimpleComment
	Author publicapi.Author
}

type frontendCommentPaginationResponse struct {
	Comments []frontendCommentResponse `json:"comments"`
	Paging   publicapi.CursorPaging    `json:"paging"`
}
