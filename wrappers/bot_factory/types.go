package bot_factory

type SetSuperInfluencerRequest struct {
	UserId     int64   `json:"user_id"`
	ContentIds []int64 `json:"content_ids"`
}

type SetSuperInfluencerResponse struct {
	UserId int64 `json:"user_id"`
}
