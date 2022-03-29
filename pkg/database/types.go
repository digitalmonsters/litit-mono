package database

import (
	"database/sql/driver"
	"encoding/json"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/google/uuid"
	"gopkg.in/guregu/null.v4"
	"time"
)

type Notification struct {
	Id                   uuid.UUID                    `json:"id" gorm:"primaryKey;autoIncrement"`
	UserId               int64                        `json:"user_id"`
	Type                 string                       `json:"type"`
	Title                string                       `json:"title"`
	Message              string                       `json:"message"`
	RelatedUserId        null.Int                     `json:"related_user_id"`
	CommentId            null.Int                     `json:"comment_id"`
	Comment              *NotificationComment         `json:"comment" gorm:"type:jsonb"`
	ContentId            null.Int                     `json:"content_id"`
	Content              *NotificationContent         `json:"content" gorm:"type:jsonb"`
	QuestionId           null.Int                     `json:"question_id"`
	CreatedAt            time.Time                    `json:"created_at"`
	KycReason            *eventsourcing.KycReason     `json:"kyc_reason"`
	KycStatus            *eventsourcing.KycStatusType `json:"kyc_status"`
	ContentCreatorStatus *eventsourcing.CreatorStatus `json:"content_creator_status"`
}

func (Notification) TableName() string {
	return "notifications"
}

type NotificationContent struct {
	Id      int64  `json:"id"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	VideoId string `json:"video_id"`
}

func (n *NotificationContent) Scan(input interface{}) error {
	return json.Unmarshal(input.([]byte), n)
}

func (n NotificationContent) Value() (driver.Value, error) {
	return json.Marshal(n)
}

type NotificationCommentType int

const (
	NotificationCommentTypeProfile = NotificationCommentType(0)
	NotificationCommentTypeContent = NotificationCommentType(1)
)

type NotificationComment struct {
	Id        int64                   `json:"id"`
	Type      NotificationCommentType `json:"type"`
	Comment   string                  `json:"comment"`
	ParentId  null.Int                `json:"parent_id"`
	ContentId null.Int                `json:"content_id"`
	ProfileId null.Int                `json:"profile_id"`
}

func (n *NotificationComment) Scan(input interface{}) error {
	return json.Unmarshal(input.([]byte), n)
}

func (n NotificationComment) Value() (driver.Value, error) {
	return json.Marshal(n)
}

type UserNotification struct {
	UserId      int64 `json:"user_id"`
	UnreadCount int64 `json:"unread_count"`
}

func (UserNotification) TableName() string {
	return "user_notifications"
}
