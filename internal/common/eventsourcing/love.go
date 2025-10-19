package eventsourcing

import "fmt"

type UserContentLoveEventData struct {
	UserId          int64 `json:"user_id"`
	ContentId       int64 `json:"content_id"`
	Love            bool  `json:"love"`
	ContentAuthorId int64 `json:"content_author_id"`
}

func (l UserContentLoveEventData) GetPublishKey() string {
	return fmt.Sprintf("%v_%v", l.UserId, l.ContentId)
}

type ContentLoveEventData struct {
	Id    int64 `json:"id"`
	Loves int64 `json:"loves"`
}

func (l ContentLoveEventData) GetPublishKey() string {
	return fmt.Sprintf("%v", l.Id)
}

type UserLoveEventData struct {
	UserId int64 `json:"user_id"`
	Loves  int64 `json:"loves"`
}

func (l UserLoveEventData) GetPublishKey() string {
	return fmt.Sprintf("%v", l.UserId)
}
