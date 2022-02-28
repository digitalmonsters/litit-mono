package eventsourcing

import (
	"fmt"
	"time"
)

type UserTotalWatchTimeEvent struct { // local.total_watch_time
	UserId         int64     `json:"user_id"`
	TotalWatchTime int64     `json:"total_watch_time"`
	CreatedAt      time.Time `json:"created_at"`
}

func (v UserTotalWatchTimeEvent) GetPublishKey() string {
	return fmt.Sprintf("%v", v.UserId)
}

type UserTodayWatchTimeEvent struct { // local.today_watch_time
	UserId              int64     `json:"user_id"`
	TotalTodayWatchTime int64     `json:"total_today_watch_time"`
	CreatedAt           time.Time `json:"created_at"`
}

func (v UserTodayWatchTimeEvent) GetPublishKey() string {
	return fmt.Sprintf("%v", v.UserId)
}
