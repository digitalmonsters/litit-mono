package eventsourcing

import "fmt"

type UserContentDislikeEventData struct {
	UserId          int64 `json:"user_id"`
	ContentId       int64 `json:"content_id"`
	Dislike         bool  `json:"dislike"`
	ContentAuthorId int64 `json:"content_author_id"`
}

func (l UserContentDislikeEventData) GetPublishKey() string {
	return fmt.Sprintf("%v_%v", l.UserId, l.ContentId)
}

type ContentDislikeEventData struct {
	Id       int64 `json:"id"`
	Dislikes int64 `json:"dislikes"`
}

func (l ContentDislikeEventData) GetPublishKey() string {
	return fmt.Sprintf("%v", l.Id)
}

type UserDislikeEventData struct {
	UserId   int64 `json:"user_id"`
	Dislikes int64 `json:"dislikes"`
}

func (l UserDislikeEventData) GetPublishKey() string {
	return fmt.Sprintf("%v", l.UserId)
}
