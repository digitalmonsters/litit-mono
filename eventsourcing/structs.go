package eventsourcing

import (
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
	return fmt.Sprintf("%v_%v", l.UserId, l.ContentId)
}

type UserCategoryEvent struct {
	UserId     int64     `json:"user_id"`
	CategoryId int64     `json:"category_id"`
	Subscribed bool      `json:"subscribed"`
	CreatedAt  time.Time `json:"created_at"`
}

func (l UserCategoryEvent) GetPublishKey() string {
	return fmt.Sprintf("%v_%v", l.UserId, l.CategoryId)
}

type UserHashtagEvent struct {
	UserId     int64  `json:"user_id"`
	Hashtag    string `json:"hashtag"`
	Subscribed bool   `json:"subscribed"`
}

func (l UserHashtagEvent) GetPublishKey() string {
	return fmt.Sprintf("%v_%v", l.UserId, l.Hashtag)
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
	return fmt.Sprintf("%v_%v", l.UserId, l.ContentId)
}

type FollowEvent struct {
	ToUserId  int64     `json:"to_user_id"`
	UserId    int64     `json:"user_id"`
	Followed  bool      `json:"followed"`
	CreatedAt time.Time `json:"created_at"`
}

func (l FollowEvent) GetPublishKey() string {
	return fmt.Sprintf("%v_%v", l.UserId, l.ToUserId)
}
