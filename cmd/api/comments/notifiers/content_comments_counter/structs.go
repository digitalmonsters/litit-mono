package content_comments_counter

import (
	"fmt"
)

type eventData struct {
	ContentId int64 `json:"content_id"`
	Count     int64 `json:"count"`
}

func (l eventData) GetPublishKey() string {
	return fmt.Sprintf("%v", l.ContentId)
}
