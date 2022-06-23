package settings

import "github.com/digitalmonsters/notification-handler/pkg/database"

type GetSettingsResponse struct {
	Settings map[string]bool `json:"settings"`
}

type ChangeSettingsRequest struct {
	Settings map[string]bool `json:"settings"`
}

type GetPushSettingsByAdminRequest struct {
	UserId int64 `json:"user_id"`
}

type GetPushSettingsByAdminItem struct {
	database.RenderTemplate
	Muted bool `json:"muted"`
}

type ChangePushSettingsByAdminRequest struct {
	ChangeSettingsRequest
	UserId int64 `json:"user_id"`
}
