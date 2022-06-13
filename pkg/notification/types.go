package notification

import (
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/google/uuid"
	"gopkg.in/guregu/null.v4"
	"time"
)

type TypeGroup string

const (
	TypeGroupAll       = "all"
	TypeGroupComment   = "comment"
	TypeGroupSystem    = "system"
	TypeGroupFollowing = "following"
)

type NotificationsResponse struct {
	Data        []NotificationsResponseItem `json:"data"`
	Next        string                      `json:"next"`
	Prev        string                      `json:"prev"`
	UnreadCount int64                       `json:"unreadCount"`
}

type NotificationsResponseItem struct {
	Id                   uuid.UUID                     `json:"id"`
	UserId               int64                         `json:"user_id"`
	Type                 string                        `json:"type"`
	Title                string                        `json:"title"`
	Message              string                        `json:"message"`
	RelatedUserId        null.Int                      `json:"related_user_id"`
	RelatedUser          *NotificationsResponseUser    `json:"related_user"`
	RenderingVariables   database.RenderingVariables   `json:"rendering_variables"`
	CustomData           database.CustomData           `json:"custom_data"`
	CommentId            null.Int                      `json:"comment_id"`
	Comment              *database.NotificationComment `json:"comment"`
	ContentId            null.Int                      `json:"content_id"`
	Content              *NotificationsResponseContent `json:"content"`
	QuestionId           null.Int                      `json:"question_id"`
	KycStatus            *eventsourcing.KycStatusType  `json:"kyc_status"`
	ContentCreatorStatus *user_go.CreatorStatus        `json:"content_creator_status"`
	KycReason            *eventsourcing.KycReason      `json:"kyc_reason,omitempty"`
	CreatedAt            time.Time                     `json:"created_at"`
}

type NotificationsResponseContent struct {
	database.NotificationContent
	ThumbUrl string `json:"thumb_url"`
}

type NotificationsResponseUser struct {
	Id                int64                     `json:"id"`
	Username          null.String               `json:"username"`
	Firstname         string                    `json:"firstname"`
	Lastname          string                    `json:"lastname"`
	Deleted           bool                      `json:"deleted"`
	Verified          bool                      `json:"verified"`
	IsBlocked         bool                      `json:"is_blocked"`
	IsFollowing       bool                      `json:"is_following"`
	IsFollower        bool                      `json:"is_follower"`
	AvatarUrl         null.String               `json:"avatar_url"`
	NamePrivacyStatus user_go.NamePrivacyStatus `json:"name_privacy_status"`
}

type Sorting struct {
	Field       string `json:"field"`
	IsAscending bool   `json:"is_ascending"`
}

type ListNotificationsByAdminRequest struct {
	Limit   int       `json:"limit"`
	Offset  int       `json:"offset"`
	Sorting []Sorting `json:"sorting"`
}

type ListNotificationsByAdminResponse struct {
	Items      []NotificationsResponseItem `json:"items"`
	TotalCount null.Int                    `json:"total_count"`
}
