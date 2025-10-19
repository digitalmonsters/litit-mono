package eventsourcing

import (
	"fmt"
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
	Active       bool      `json:"active"`
	Comment      string    `json:"comment"`
	ContentId    null.Int  `json:"content_id"`
	ParentId     null.Int  `json:"parent_id"`
	ProfileId    null.Int  `json:"profile_id"`
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

type CommentCountOnContentEvent struct {
	ContentId int64 `json:"content_id"`
	Count     int64 `json:"count"`
}

func (l CommentCountOnContentEvent) GetPublishKey() string {
	return fmt.Sprintf("%v", l.ContentId)
}
