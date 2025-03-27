package notification_handler

import (
	"context"
	"fmt"
	"time"

	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/wrappers"
)

type INotificationHandlerWrapper interface {
	EnqueueNotificationWithTemplate(templateName string, userId int64,
		renderingVars map[string]string, customData map[string]interface{}, ctx context.Context) chan EnqueueMessageResult
	EnqueueNotificationWithCustomTemplate(title, body, headline string, userId int64, customData map[string]interface{}, ctx context.Context) chan EnqueueMessageResult
	GetNotificationsReadCount(notificationIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]int64]
	DisableUnregisteredTokens(tokens []string, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[[]string]
	CreateNotification(notifications Notification, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[CreateNotificationResponse]
	DeleteNotificationByIntroID(introID int, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[DeleteNotificationByIntroIDResponse]
}

//goland:noinspection GoNameStartsWithPackageName
type NotificationHandlerWrapper struct {
	baseWrapper     *wrappers.BaseWrapper
	defaultTimeout  time.Duration
	apiUrl          string
	serviceName     string
	publisher       eventsourcing.IEventPublisher
	customPublisher eventsourcing.IEventPublisher
}

type EnqueueMessageResult struct {
	Id    string `json:"id"`
	Error error  `json:"error"`
}
type SendNotificationWithTemplate struct {
	Id                 string                 `json:"id"`
	TemplateName       string                 `json:"template_name"`
	RenderingVariables map[string]string      `json:"rendering_variables"`
	CustomData         map[string]interface{} `json:"custom_data"`
	UserId             int64                  `json:"user_id"`
}

func (s SendNotificationWithTemplate) GetPublishKey() string {
	return fmt.Sprint(s.UserId)
}

type SendNotificationWithCustomTemplate struct {
	Id         string                 `json:"id"`
	UserId     int64                  `json:"user_id"`
	Title      string                 `json:"title"`
	Body       string                 `json:"body"`
	Headline   string                 `json:"headline"`
	CustomData map[string]interface{} `json:"custom_data"`
}

func (s SendNotificationWithCustomTemplate) GetPublishKey() string {
	return fmt.Sprint(s.UserId)
}

type NotificationChannel int

const (
	NotificationChannelPush  = NotificationChannel(1)
	NotificationChannelEmail = NotificationChannel(2)
)

type GetNotificationsReadCountRequest struct {
	NotificationIds []int64 `json:"notification_ids"`
}

type DisableUnregisteredTokensRequest struct {
	Tokens []string `json:"tokens"`
}

type Notification struct {
	UserID               int                    `gorm:"not null"`
	Type                 string                 `gorm:"type:varchar(255);not null"`
	Title                string                 `gorm:"type:varchar(255);not null"`
	Message              string                 `gorm:"type:varchar(255);not null"`
	RelatedUserID        *int                   `gorm:"type:int"`
	CommentID            *int                   `gorm:"type:int"`
	Comment              map[string]interface{} `gorm:"type:jsonb"`
	ContentID            *int                   `gorm:"type:int"`
	Content              map[string]interface{} `gorm:"type:jsonb"`
	QuestionID           *int                   `gorm:"type:int"`
	CreatedAt            time.Time              `gorm:"default:CURRENT_TIMESTAMP"`
	KycReason            *string                `gorm:"type:varchar(255)"`
	KycStatus            *string                `gorm:"type:varchar(255)"`
	ContentCreatorStatus *int                   `gorm:"type:int"`
	RenderingVariables   map[string]interface{} `gorm:"type:jsonb;default:'{}'"`
	CustomData           map[string]interface{} `gorm:"type:jsonb;default:'{}'"`
	InApp                bool                   `json:"in_app"`
	CollapseKey          string                 `json:"collapse_key"`
	TriggerFireBase      bool                   `json:"trigger_firebase"`
}

type CreateNotificationRequest struct {
	Notifications Notification
}

type CreateNotificationResponse struct {
	Status bool
}

type DeleteNotificationByIntroIDRequest struct {
	IntroID int `json:"intro_id"`
}

type DeleteNotificationByIntroIDResponse struct {
	Status bool
}

type GenericEmailRPCRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type GenericHTMLEmailRPCRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type GenericEmailResponse struct {
	Status bool
}
