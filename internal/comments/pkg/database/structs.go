package database

import (
	"time"

	"gopkg.in/guregu/null.v4"
)

type Comment struct {
	Id           int64     `json:"id"`
	AuthorId     int64     `json:"author_id"`
	NumReplies   int64     `json:"num_replies"`
	NumUpvotes   int64     `json:"num_upvotes"`
	NumDownvotes int64     `json:"num_downvotes"`
	CreatedAt    time.Time `json:"created_at"`
	Active       bool      `json:"active"`
	Comment      string    `json:"comment"`
	ContentId    null.Int  `json:"content_id"`
	ProfileId    null.Int  `json:"profile_id"`
	ParentId     null.Int  `json:"parent_id"`
	NumReports   int64     `json:"num_reports"`
}

func (Comment) TableName() string {
	return "comment"
}

type CommentVote struct {
	UserId    int64     `json:"user_id" gorm:"primaryKey"`
	CommentId int64     `json:"comment_id" gorm:"primaryKey"`
	VoteUp    null.Bool `json:"vote_up"`
}

func (CommentVote) TableName() string {
	return "comment_vote"
}

type Report struct {
	Id         int
	ContentId  null.Int
	UserId     null.Int
	Type       string
	ReportType string
	ReporterId int64
	CommentId  int64
	Detail     string
	CreatedAt  time.Time
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
