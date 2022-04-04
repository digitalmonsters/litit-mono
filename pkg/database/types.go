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
	Upvote    null.Bool               `json:"upvote"`
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

func GetNotificationType(templateId string) string{
	switch templateId {
	case "first_daily_followers_bonus":
		return "push.bonus.daily_followers.first"
	case "first_daily_time_bonus":
		return "push.bonus.daily_time.first"
	case "first_guest_x_earned_points":
		return "push.earned_points.first"
	case "first_guest_x_paid_views":
		return "push.paid_views.first"
	case "first_referral_joined":
		return "push.referral.first"
	case "first_video_shared":
		return "push.share.first"
	case "first_weekly_followers_bonus":
		return "push.bonus.weekly_followers.first"
	case "first_weekly_time_bonus":
		return "push.bonus.weekly_time.first"
	case "first_x_paid_views":
		return "push.paid_views.first"
	case "first_x_paid_views_as_content_owner":
		return "push.content_owner.paid_views.first"
	case "guest_max_earned_points_for_views":
		return "push.earned_points.max"
	case "increase_reward_stage_1":
		return "push.referral.reward_increase.stage1"
	case "increase_reward_stage_2":
		return "push.referral.reward_increase.stage2"
	case "registration_verify_bonus":
		return "push.bonus.registration.verify"
	case "other_referrals_joined":
		return "push.referral.other"
	case "custom_reward_increase":
		return "push.referral.reward_increase"
	}
	return ""
}

func GetMarketingNotifications() []string{
	return []string{
		"push.bonus.daily_followers.first", "push.bonus.daily_time.first", "push.earned_points.first",
		"push.paid_views.first", "push.referral.first", "push.share.first", "push.bonus.weekly_followers.first",
		"push.bonus.weekly_time.first", "push.content_owner.paid_views.first", "push.earned_points.max",
		"push.referral.reward_increase.stage1", "push.referral.reward_increase.stage2", "push.bonus.registration.verify",
		"push.referral.other", "push.referral.reward_increase",
	}
}