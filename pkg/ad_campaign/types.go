package ad_campaign

import (
	"github.com/digitalmonsters/ads-manager/pkg/database"
	"gopkg.in/guregu/null.v4"
)

type CreateAdCampaignRequest struct {
	Name         string          `json:"name"`
	AdType       database.AdType `json:"ad_type"`
	ContentId    int64           `json:"content_id"`
	Link         null.String     `json:"link"`
	LinkButtonId null.Int        `json:"link_button_id"`
	Country      null.String     `json:"country"`
	DurationMin  uint            `json:"duration_min"`
	Budget       uint            `json:"budget"`
	Gender       null.String     `json:"gender"`
	AgeFrom      uint            `json:"age_from"`
	AgeTo        uint            `json:"age_to"`
}
