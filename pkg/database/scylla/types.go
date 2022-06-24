package scylla

import (
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
