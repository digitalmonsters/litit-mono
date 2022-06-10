package eventsourcing

import "fmt"

type AdminPushMessageEventData struct {
	UserId     int64                  `json:"user_id"`
	Title      string                 `json:"title"`
	Message    string                 `json:"message"`
	CustomData map[string]interface{} `json:"custom_data"`
}

func (t AdminPushMessageEventData) GetPublishKey() string {
	return fmt.Sprintf("%v", t.UserId)
}
