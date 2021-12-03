package database

import (
	"gopkg.in/guregu/null.v4"
	"time"
)

type Comment struct {
	Id           int64     `json:"id"`
	AuthorId     int64     `json:"author_id"`
	NumReplies   int64     `json:"num_replies"`
	NumUpvotes   int64     `json:"num_upvotes"`
	NumDownvotes int64     `json:"num_downvotes"`
	CreatedAt    time.Time `json:"created_at"`
	Active       bool
	Comment      string   `json:"comment"`
	ContentId    int64    `json:"content_id"`
	ParentId     null.Int `json:"parent_id"`
}

func (Comment) TableName() string {
	return "comment"
}

type CommentVote struct {
}

func (CommentVote) TableName() string {
	return "comment_vote"
}

type CommentForDelete struct {
	Id         int64    `json:"id"`
	AuthorId   int64    `json:"author_id"`
	NumReplies int64    `json:"num_replies"`
	ContentId  int64    `json:"content_id"`
	ParentId   null.Int `json:"parent_id"`
}

func (CommentForDelete) TableName() string {
	return "comment"
}

type CommentWithAuthorId struct {
	Id       int64  `json:"id"`
	AuthorId int64  `json:"author_id"`
	Comment  string `json:"comment"`
}

func (CommentWithAuthorId) TableName() string {
	return "comment"
}
