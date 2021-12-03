package publicapi

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
	Author       Author    `json:"author"`
}

type Author struct {
	Id        int64       `json:"id"`
	Username  string      `json:"username"`
	Avatar    null.String `json:"avatar"`
	Firstname string      `json:"firstname"`
	Lastname  string      `json:"lastname"`
}
