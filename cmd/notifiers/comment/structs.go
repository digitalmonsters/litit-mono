package comment

import (
	"fmt"
	"gopkg.in/guregu/null.v4"
	"time"
)

type eventData struct {
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
	Width           null.Int    `json:"width"`
	Height          null.Int    `json:"height"`
	VideoId         null.String `json:"video_id"`
	ContentAuthorId null.Int    `json:"content_author_id"`
	ProfileId       null.Int    `json:"profile_id"`
	EventType       EventType   `json:"event_type"`
}

func (l eventData) GetPublishKey() string {
	return fmt.Sprintf("%v", l.Id)
}

type EventType int

const (
	ContentResourceTypeCreate EventType = 1
	ContentResourceTypeUpdate EventType = 2
	ContentResourceTypeDelete EventType = 3
	ProfileResourceTypeCreate EventType = 4
	ProfileResourceTypeUpdate EventType = 5
	ProfileResourceTypeDelete EventType = 6
)
