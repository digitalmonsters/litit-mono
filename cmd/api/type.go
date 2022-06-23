package api

import (
	"github.com/digitalmonsters/notification-handler/pkg/database/scylla"
	"time"
)

type pingRequest struct {
	Data string `json:"data"`
}

type GeneralPushNotificationTaskRequest struct {
	CurrentDate time.Time `json:"current_date"`
}

type UserPushNotificationTaskRequest struct {
	Item        scylla.PushNotificationGroupQueue `json:"item"`
	CurrentDate time.Time                         `json:"current_date"`
}
