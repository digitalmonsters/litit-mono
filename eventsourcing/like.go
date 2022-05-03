package eventsourcing

import "fmt"

type UserContentEventData struct {
	UserId          int64 `json:"user_id"`
	ContentId       int64 `json:"content_id"`
	Like            bool  `json:"like"`
	ContentAuthorId int64 `json:"content_author_id"`
}

func (l UserContentEventData) GetPublishKey() string {
	return fmt.Sprintf("%v_%v", l.UserId, l.ContentId)
}

type ContentLikeEventData struct {
	Id    int64 `json:"id"`
	Likes int64 `json:"likes"`
}

func (l ContentLikeEventData) GetPublishKey() string {
	return fmt.Sprintf("%v", l.Id)
}

type UserLikeEventData struct {
	UserId int64 `json:"user_id"`
	Likes  int64 `json:"likes"`
}

func (l UserLikeEventData) GetPublishKey() string {
	return fmt.Sprintf("%v", l.UserId)
}
