package database

import (
	"database/sql/driver"
	"encoding/json"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
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
	ContentCreatorStatus *user_go.CreatorStatus       `json:"content_creator_status"`
	RenderingVariables   RenderingVariables           `json:"rendering_variables"`
	CustomData           CustomData                   `json:"custom_data"`
}

type RenderingVariables map[string]string

func (n *RenderingVariables) Scan(input interface{}) error {
	return json.Unmarshal(input.([]byte), n)
}

func (n RenderingVariables) Value() (driver.Value, error) {
	return json.Marshal(n)
}

type CustomData map[string]interface{}

func (n *CustomData) Scan(input interface{}) error {
	return json.Unmarshal(input.([]byte), n)
}

func (n CustomData) Value() (driver.Value, error) {
	return json.Marshal(n)
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

func GetNotificationType(templateId string) string {
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
	case "first_time_avatar_added":
		return "push.avatar.first"
	case "add_description_bonus":
		return "push.description.first"
	case "first_video_uploaded":
		return "push.upload.first"
	case "first_spot_uploaded":
		return "push.upload.spot.first"
	case "user_need_to_first_upload":
		return "push.user.need.upload"
	case "user_need_to_upload_avatar":
		return "push.user.need.avatar"
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
	case "megabonus":
		return "push.referral.megabonus"
	case "guest_after_install_first_push":
		return "push.guest.after_install"
	case "guest_after_install_second_push":
		return "push.guest.after_install"
	case "guest_after_install_third_push":
		return "push.guest.after_install"
	case "user_after_signup_first_push":
		return "push.user.after_signup"
	case "user_after_signup_second_push":
		return "push.user.after_signup"
	case "user_after_signup_third_push":
		return "push.user.after_signup"
	case "user_after_signup_fourth_push":
		return "push.user.after_signup"
	case "user_after_signup_fifth_push":
		return "push.user.after_signup"
	case "daily_max_amount_of_paid_views_reached":
		return "push.user.paid_views.daily_max"
	case "daily_max_amount_of_paid_spot_views_reached":
		return "push.user.paid_spot_views.daily_max"
	case "first_x_paid_views_gender_push":
		return "push.gender.first_x_paid_views"
	case "first_email_marketing_added":
		return "push.user.first_email_marketing_added"
	}
	return ""
}

//without push.user.after_signup, push.guest.after_install
func GetMarketingNotifications() []string {
	return []string{
		"push.bonus.daily_followers.first", "push.bonus.daily_time.first", "push.earned_points.first",
		"push.paid_views.first", "push.referral.first", "push.share.first", "push.bonus.weekly_followers.first",
		"push.bonus.weekly_time.first", "push.content_owner.paid_views.first", "push.earned_points.max",
		"push.referral.reward_increase.stage1", "push.referral.reward_increase.stage2", "push.bonus.registration.verify",
		"push.referral.other", "push.referral.reward_increase", "push.referral.megabonus",
		"push.avatar.first", "push.upload.first", "push.upload.spot.first", "push.description.first", "push.gender.first_x_paid_views",
		"push.user.first_email_marketing_added",
	}
}
