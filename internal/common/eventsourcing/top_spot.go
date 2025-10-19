package eventsourcing

import (
	"fmt"
	"time"
)

type TopSpotEventData struct {
	UserId    int64            `json:"user_id"`
	ContentId int64            `json:"content_id"`
	Type      TopSpotEventType `json:"type"`
	CreatedAt time.Time        `json:"created_at"`
}

type TopSpotEventType int

const (
	TopSpotEventTypeDaily  = TopSpotEventType(1)
	TopSpotEventTypeWeekly = TopSpotEventType(2)
)

func (t TopSpotEventData) GetPublishKey() string {
	return fmt.Sprintf("%v_%v", t.UserId, t.ContentId)
}
