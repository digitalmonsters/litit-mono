package ad_campaign

import (
	"github.com/digitalmonsters/ads-manager/pkg/database"
	"gopkg.in/guregu/null.v4"
)

type CreateAdCampaignRequest struct {
	Name         string
	AdType       database.AdType
	ContentId    int64
	Link         null.String
	LinkButtonId null.Int
	Country      null.String
	DurationMin  uint
	Budget       uint
	Gender       null.String
	AgeFrom      uint
	AgeTo        uint
}
