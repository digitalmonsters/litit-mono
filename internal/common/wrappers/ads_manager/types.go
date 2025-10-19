package ads_manager

import (
	"gopkg.in/guregu/null.v4"
)

type GetAdsContentForUserRequest struct {
	UserId             int64   `json:"user_id"`
	ContentIdsToMix    []int64 `json:"content_ids_to_mix"`
	ContentIdsToIgnore []int64 `json:"content_ids_to_ignore"`
}

type GetAdsContentForUserResponse struct {
	MixedContentIdsWithAd []int64              `json:"mixed_content_ids_with_ad"`
	ContentAds            map[int64]*ContentAd `json:"content_ads"`
}

type ContentAd struct {
	ContentId      int64       `json:"content_id"`
	Link           null.String `json:"link"`
	LinkButtonId   null.Int    `json:"link_button_id"`
	LinkButtonName null.String `json:"link_button_name"`
}
