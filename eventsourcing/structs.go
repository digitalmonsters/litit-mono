package eventsourcing

import (
	"encoding/json"
	"fmt"
	"gopkg.in/guregu/null.v4"
	"time"
)

type LikeEvent struct {
	UserId    int64     `json:"user_id"`
	ContentId int64     `json:"content_id"`
	Like      bool      `json:"like"`
	CreatedAt time.Time `json:"created_at"`
}

func (l LikeEvent) GetPublishKey() string {
	return fmt.Sprintf("{\"content_id\":%v,\"user_id\":%v}", l.ContentId, l.UserId)
}

type UserCategoryEvent struct {
	UserId     int64     `json:"user_id"`
	CategoryId int64     `json:"category_id"`
	Subscribed bool      `json:"subscribed"`
	CreatedAt  time.Time `json:"created_at"`
}

func (l UserCategoryEvent) GetPublishKey() string {
	return fmt.Sprintf("{\"category_id\":%v,\"user_id\":%v}", l.CategoryId, l.UserId)
}

type UserHashtagEvent struct {
	UserId     int64  `json:"user_id"`
	Hashtag    string `json:"hashtag"`
	Subscribed bool   `json:"subscribed"`
}

func (l UserHashtagEvent) GetPublishKey() string {
	name := l.Hashtag
	if v, _ := json.Marshal(name); len(v) > 0 {
		name = string(v)
	}
	return fmt.Sprintf("{\"hashtag\":\"%v\",\"user_id\":%v}", name, l.UserId)
}

type ViewEvent struct {
	UserId    int64       `json:"user_id"`
	ContentId int64       `json:"content_id"`
	Duration  int         `json:"duration"`
	UserIp    string      `json:"user_ip"`
	SharerId  null.Int    `json:"sharer_id"`
	ShareCode null.String `json:"share_code"`
	AdsId     null.Int    `json:"ads_id"`
	CreatedAt time.Time   `json:"created_at"`
}

func (l ViewEvent) GetPublishKey() string {
	return fmt.Sprintf("{\"content_id\":%v,\"user_id\":%v}", l.ContentId, l.UserId)
}

type FollowEvent struct {
	ToUserId  int64     `json:"to_user_id"`
	UserId    int64     `json:"user_id"`
	Followed  bool      `json:"followed"`
	CreatedAt time.Time `json:"created_at"`
}

func (l FollowEvent) GetPublishKey() string {
	return fmt.Sprintf("{\"user_id\":%v,\"to_user_id\":%v}", l.UserId, l.ToUserId)
}
