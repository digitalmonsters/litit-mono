package eventsourcing

import (
	"fmt"
	"time"
)

type TodayFollowersData struct {
	UserId         int64     `json:"user_id"`
	TodayFollowers int64     `json:"today_followers"`
	CreatedAt      time.Time `json:"created_at"`
}

func (l TodayFollowersData) GetPublishKey() string {
	return fmt.Sprintf("%v", l.UserId)
}
