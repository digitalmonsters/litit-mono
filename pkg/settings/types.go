package settings

type GetSettingsResponse struct {
	Settings map[string]bool `json:"settings"`
}

type ChangeSettingsRequest struct {
	Settings map[string]bool `json:"settings"`
}
