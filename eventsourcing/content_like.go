package eventsourcing

import "fmt"

// USER_CONTENT_LIKES
type UserContentEventData struct {
	UserId          int64 `json:"user_id"`
	ContentId       int64 `json:"content_id"`
	Like            bool  `json:"like"`
	ContentAuthorId int64 `json:"content_author_id"`
}

func (l UserContentEventData) GetPublishKey() string {
	return fmt.Sprintf("%v", l.UserId)
}
