package scylla

import (
	"github.com/digitalmonsters/go-common/translation"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"time"
)

type Notification struct {
	UserId             int64     `json:"user_id"`
	EventType          string    `json:"event_type"`
	EntityId           int64     `json:"entity_id"`
	RelatedEntityId    int64     `json:"related_entity_id"`
	CreatedAt          time.Time `json:"created_at"`
	NotificationsCount int64     `json:"notifications_count"`

	Title              string `json:"title"`
	Body               string `json:"body"`
	Headline           string `json:"headline"`
	Kind               string `json:"kind"`
	RenderingVariables string `json:"rendering_variables"`
	CustomData         string `json:"custom_data"`
	NotificationInfo   string `json:"notification_info"`
}

type NotificationRelation struct {
	UserId          int64  `json:"user_id"`
	EventType       string `json:"event_type"`
	EntityId        int64  `json:"entity_id"`
	RelatedEntityId int64  `json:"related_entity_id"`
	EventApplied    bool   `json:"event_applied"`
}

type PushNotificationGroupQueue struct {
	DeadlineKey       time.Time `json:"deadline_key"`
	Deadline          time.Time `json:"deadline"`
	UserId            int64     `json:"user_id"`
	EventType         string    `json:"event_type"`
	EntityId          int64     `json:"entity_id"`
	CreatedAt         time.Time `json:"created_at"`
	NotificationCount int64     `json:"notification_count"`
}

type NotificationByTypeGroup struct {
	UserId          int64     `json:"user_id"`
	EventType       string    `json:"event_type"`
	EntityId        int64     `json:"entity_id"`
	RelatedEntityId int64     `json:"related_entity_id"`
	CreatedAt       time.Time `json:"created_at"`
}

type User struct {
	ClusterKey        string                    `json:"cluster_key"`
	UserId            int64                     `json:"user_id"`
	Username          string                    `json:"username"`
	Firstname         string                    `json:"firstname"`
	Lastname          string                    `json:"lastname"`
	NamePrivacyStatus user_go.NamePrivacyStatus `json:"name_privacy_status"`
	Language          translation.Language      `json:"language"`
	Email             string                    `json:"email"`
}

func GetUserClusterKey(userId int64) int64 {
	return userId / 30000
}
