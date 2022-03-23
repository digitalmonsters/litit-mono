package eventsourcing

import (
	"fmt"
	"gopkg.in/guregu/null.v4"
	"time"
)

type Comment struct {
	Id              int64       `json:"id"`
	AuthorId        int64       `json:"author_id"`
	NumReplies      int64       `json:"num_replies"`
	NumUpvotes      int64       `json:"num_upvotes"`
	NumDownvotes    int64       `json:"num_downvotes"`
	CreatedAt       time.Time   `json:"created_at"`
	Active          bool        `json:"active"`
	Comment         string      `json:"comment"`
	ContentId       null.Int    `json:"content_id"`
	ParentId        null.Int    `json:"parent_id"`
	ParentAuthorId  null.Int    `json:"parent_author_id"`
	Width           null.Int    `json:"width"`
	Height          null.Int    `json:"height"`
	VideoId         null.String `json:"video_id"`
	ContentAuthorId null.Int    `json:"content_author_id"`
	ProfileId       null.Int    `json:"profile_id"`
	BaseChangeEvent
}

func (l Comment) GetPublishKey() string {
	return fmt.Sprintf("%v", l.Id)
}

type CommentChangeReason int

const (
	CommentChangeReasonContent CommentChangeReason = 1
	CommentChangeReasonProfile CommentChangeReason = 2
)

type Vote struct {
	UserId          int64     `json:"user_id"`
	Upvote          null.Bool `json:"upvote"`
	CommentId       int64     `json:"comment_id"`
	ParentId        null.Int  `json:"parent_id"`
	CommentAuthorId int64     `json:"comment_author_id"`
	Comment         string    `json:"comment"`
	EntityId        int64     `json:"entity_id"`
	BaseChangeEvent
}

func (l Vote) GetPublishKey() string {
	return fmt.Sprintf("%v_%v", l.CommentId, l.UserId)
}
