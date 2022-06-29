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
	case "top_daily_spot_bonus":
		return "push.user.daily_top_spot_reward"
	case "top_weekly_spot_bonus":
		return "push.user.weekly_top_spot_reward"
	case "last_boring_spots":
		return "push.user.boring_spots"
	case "first_boring_spots":
		return "push.user.boring_spots"
	case "warning_boring_spots":
		return "push.user.boring_spots"
	}
	return ""
}

func GetNotificationTemplates(notificationType string) []string {
	switch notificationType {
	case "push.bonus.daily_followers.first":
		return []string{"first_daily_followers_bonus"}
	case "push.bonus.daily_time.first":
		return []string{"first_daily_time_bonus"}
	case "push.earned_points.first":
		return []string{"first_guest_x_earned_points"}
	case "push.paid_views.first":
		return []string{"first_guest_x_paid_views", "first_x_paid_views"}
	case "push.referral.first":
		return []string{"first_referral_joined"}
	case "push.share.first":
		return []string{"first_video_shared"}
	case "push.bonus.weekly_followers.first":
		return []string{"first_weekly_followers_bonus"}
	case "push.bonus.weekly_time.first":
		return []string{"first_weekly_time_bonus"}
	case "push.avatar.first":
		return []string{"first_time_avatar_added"}
	case "push.description.first":
		return []string{"add_description_bonus"}
	case "push.upload.first":
		return []string{"first_video_uploaded"}
	case "push.upload.spot.first":
		return []string{"first_spot_uploaded"}
	case "push.user.need.upload":
		return []string{"user_need_to_first_upload"}
	case "push.user.need.avatar":
		return []string{"user_need_to_upload_avatar"}
	case "push.content_owner.paid_views.first":
		return []string{"first_x_paid_views_as_content_owner"}
	case "push.earned_points.max":
		return []string{"guest_max_earned_points_for_views"}
	case "push.referral.reward_increase.stage1":
		return []string{"increase_reward_stage_1"}
	case "push.referral.reward_increase.stage2":
		return []string{"increase_reward_stage_2"}
	case "push.bonus.registration.verify":
		return []string{"registration_verify_bonus"}
	case "push.referral.other":
		return []string{"other_referrals_joined"}
	case "push.referral.reward_increase":
		return []string{"custom_reward_increase"}
	case "push.referral.megabonus":
		return []string{"megabonus"}
	case "push.guest.after_install":
		return []string{"guest_after_install_first_push", "guest_after_install_second_push", "guest_after_install_third_push"}
	case "push.user.after_signup":
		return []string{"user_after_signup_first_push", "user_after_signup_second_push", "user_after_signup_third_push",
			"user_after_signup_fourth_push", "user_after_signup_fifth_push"}
	case "push.user.paid_views.daily_max":
		return []string{"daily_max_amount_of_paid_views_reached"}
	case "push.user.paid_spot_views.daily_max":
		return []string{"daily_max_amount_of_paid_spot_views_reached"}
	case "push.gender.first_x_paid_views":
		return []string{"first_x_paid_views_gender_push"}
	case "push.user.first_email_marketing_added":
		return []string{"first_email_marketing_added"}
	case "push.profile.following":
		return []string{"follow"}
	case "push.comment.reply":
		return []string{"comment_reply"}
	case "push.comment.vote":
		return []string{"comment_vote_like", "comment_vote_dislike"}
	case "push.profile.comment":
		return []string{"comment_profile_resource_create"}
	case "push.content.comment":
		return []string{"comment_content_resource_create"}
	case "push.admin.bulk":
		return []string{"push_admin"}
	case "push.content.new-posted":
		return []string{"content_posted"}
	case "push.tip":
		return []string{"tip"}
	case "push.content.like":
		return []string{"content_like"}
	case "push.bonus.followers":
		return []string{"bonus_followers"}
	case "push.bonus.daily":
		return []string{"bonus_time"}
	case "push.content.successful-upload":
		return []string{"content_upload"}
	case "push.spot.successful-upload":
		return []string{"spot_upload"}
	case "push.content.rejected":
		return []string{"content_reject"}
	case "push.kyc.status":
		return []string{"kyc_status_verified", "kyc_status_rejected"}
	case "push.content-creator.status":
		return []string{"creator_status_rejected", "creator_status_approved", "creator_status_pending"}
	case "push.user.daily_top_spot_reward":
		return []string{"top_daily_spot_bonus"}
	case "push.user.weekly_top_spot_reward":
		return []string{"top_weekly_spot_bonus"}
	case "push.user.boring_spots":
		return []string{"warning_boring_spots", "last_boring_spots", "first_boring_spots"}
	}
	return []string{}
}

type UserNotificationsSettings struct {
	ClusterKey int64
	UserId     int64
	TemplateId int64
	Enabled    bool
}

func GetUserNotificationsSettingsClusterKey(userId int64) int64 {
	return userId / 30000
}
