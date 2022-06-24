package api

import (
	"github.com/digitalmonsters/notification-handler/pkg/database/scylla"
)

type pingRequest struct {
	Data string `json:"data"`
}

type GeneralPushNotificationTaskRequest struct {
	CurrentDate string `json:"current_date"`
}

type UserPushNotificationTaskRequest struct {
	Item        scylla.PushNotificationGroupQueue `json:"item"`
	CurrentDate string                            `json:"current_date"`
}
