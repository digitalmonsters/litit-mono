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
	UserId    int64 `json:"user_id"`
	CommentId int64 `json:"comment_id"`
	VoteUp    bool  `json:"vote_up"`
}

func (CommentVote) TableName() string {
	return "comment_vote"
}

type Report struct {
	Id         int
	ContentId  int64
	Type       string
	ReporterId int64
	CommentId  int64
	Detail     string
}

func (Report) TableName() string {
	return "report"
}

type UserStatsAction struct {
	Id       int64 `json:"id"`
	Comments int64 `json:"comments"`
}

func (UserStatsAction) TableName() string {
	return "user_stats_action"
}

type UserStatsContent struct {
	Id       int64 `json:"id"`
	Comments int64 `json:"comments"`
}

func (UserStatsContent) TableName() string {
	return "user_stats_content"
}

type Content struct {
	Id            int64 `json:"id"`
	CommentsCount int64 `json:"comments_count"`
}

func (Content) TableName() string {
	return "content"
}
