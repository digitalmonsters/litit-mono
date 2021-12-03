package publicapi

import "gopkg.in/guregu/null.v4"

type GetCommentsByTypeWithResourceRequest struct {
	ContentId int64
	ParentId  int64
	After     string // cursor
	Count     int    // Limit
	SortOrder string
}

type CursorPaging struct {
	HasNext bool   `json:"hasNext"`
	Next    string `json:"next"`
}

type GetCommentsByTypeWithResourceResponse struct {
	Comments []Comment    `json:"comments"`
	Paging   CursorPaging `json:"paging"`
}

type SendCommentRequest struct {
	Comment  string   `json:"comment"`
	ParentId null.Int `json:"parent_id"`
	AuthorId int64    `json:"author_id"`
}

type SendCommentResponse struct {
	Id        int64  `json:"id"`
	Comment   string `json:"comment"`
	AuthorId  int64  `json:"author_id"`
	ContentId int64  `json:"content_id"`
}

type BlockedUserType string

const (
	BlockedUser   BlockedUserType = "BLOCKED USER"
	BlockedByUser BlockedUserType = "YOUR PROFILE IS BLOCKED BY USER"
)
