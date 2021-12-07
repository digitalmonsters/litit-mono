package comments

import (
	"gopkg.in/guregu/null.v4"
	"time"
)

type GetCommentsByTypeWithResourceRequest struct {
	ResourceId int64
	ParentId   int64
	After      string // cursor
	Before     string // cursor
	Count      int64  // Limit
	SortOrder  string
}

type CursorPaging struct {
	HasNext bool   `json:"hasNext"`
	Next    string `json:"next"`
	Before  string `json:"before"`
	After   string `json:"after"`
}

type GetCommentsByTypeWithResourceResponse struct {
	Comments []Comment    `json:"comments"`
	Paging   CursorPaging `json:"paging"`
}

type SimpleComment struct {
	Id           int64     `json:"id"`
	AuthorId     int64     `json:"author_id"`
	NumReplies   int64     `json:"num_replies"`
	NumUpvotes   int64     `json:"num_upvotes"`
	NumDownvotes int64     `json:"num_downvotes"`
	CreatedAt    time.Time `json:"created_at"`
	MyVoteUp     null.Bool `json:"my_vote_up"`
	ContentId    null.Int  `json:"content_id"`
	ProfileId    null.Int  `json:"profile_id"`
	Comment      string    `json:"comment"`
}

type Comment struct {
	SimpleComment
	Author  Author        `json:"author"`
	Content SimpleContent `json:"content"`
}

type CommentOnProfile struct {
	SimpleComment
	Author Author `json:"author"`
}

type Author struct {
	Id        int64       `json:"id"`
	Username  string      `json:"username"`
	Avatar    null.String `json:"avatar"`
	Firstname string      `json:"firstname"`
	Lastname  string      `json:"lastname"`
}

type SimpleContent struct {
	Id       int64 `json:"id"`
	AuthorId int64 `json:"author_id"`
}

type ResourceType int

const (
	ContentResourceType       ResourceType = 1
	ProfileResourceType       ResourceType = 2
	ParentCommentResourceType ResourceType = 3
)
