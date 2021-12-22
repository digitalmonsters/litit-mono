package user_comments_counter

import (
	"fmt"
)

type eventData struct {
	UserId int64 `json:"user_id"`
	Count  int64 `json:"count"`
}

func (l eventData) GetPublishKey() string {
	return fmt.Sprintf("%v", l.UserId)
}
