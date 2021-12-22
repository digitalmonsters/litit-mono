package vote

import (
	"fmt"
	"gopkg.in/guregu/null.v4"
)

type eventData struct {
	UserId          int64     `json:"user_id"`
	Upvote          null.Bool `json:"upvote"`
	CommentId       int64     `json:"comment_id"`
	ParentId        null.Int  `json:"parent_id"`
	CommentAuthorId int64     `json:"comment_author_id"`
	Comment         string    `json:"comment"`
	ContentId       null.Int  `json:"content_id"`
	ProfileId       null.Int  `json:"profile_id"`
}

func (l eventData) GetPublishKey() string {
	return fmt.Sprintf("%v_%v", l.CommentId, l.UserId)
}
