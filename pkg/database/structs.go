package database

import "time"

type Comment struct {
	Id           int64     `json:"id"`
	AuthorId     int64     `json:"author_id"`
	NumReplies   int64     `json:"num_replies"`
	NumUpvotes   int64     `json:"num_upvotes"`
	NumDownvotes int64     `json:"num_downvotes"`
	CreatedAt    time.Time `json:"created_at"`
	Active       bool
}

func (Comment) TableName() string {
	return "comment"
}

type CommentVote struct {
}

func (CommentVote) TableName() string {
	return "comment_vote"
}
