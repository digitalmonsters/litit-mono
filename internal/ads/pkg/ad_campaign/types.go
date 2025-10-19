package ad_campaign

import (
	"github.com/digitalmonsters/ads-manager/pkg/database"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/shopspring/decimal"
	"gopkg.in/guregu/null.v4"
)

type CreateAdCampaignRequest struct {
	Name          string          `json:"name"`
	AdType        database.AdType `json:"ad_type"`
	ContentId     int64           `json:"content_id"`
	Link          null.String     `json:"link"`
	LinkButtonId  null.Int        `json:"link_button_id"`
	Country       null.String     `json:"country"`
	DurationMin   uint            `json:"duration_min"`
	Budget        decimal.Decimal `json:"budget"`
	Gender        null.String     `json:"gender"`
	AgeFrom       uint            `json:"age_from"`
	AgeTo         uint            `json:"age_to"`
	CategoriesIds []int64         `json:"categories_ids"`
}

type ClickLinkRequest struct {
	ContentId int64 `json:"content_id"`
}

type StopAdCampaignRequest struct {
	AdCampaignId int64 `json:"ad_campaign_id"`
}

type StartAdCampaignRequest struct {
	AdCampaignId int64 `json:"ad_campaign_id"`
}

type ListAdCampaignsRequest struct {
	Name     null.String                `json:"name"`
	DateFrom null.Time                  `json:"date_from"`
	DateTo   null.Time                  `json:"date_to"`
	Age      null.Int                   `json:"age"`
	Status   *database.AdCampaignStatus `json:"status"`
	Limit    int                        `json:"limit"`
	Offset   int                        `json:"offset"`
}

type ListAdCampaignsResponse struct {
	Items      []*ListAdCampaignsResponseItem `json:"items"`
	TotalCount null.Int                       `json:"total_count"`
}

type ListAdCampaignsResponseItemContent struct {
	Content    content.SimpleContent
	AnimUrl    string `json:"anim_url"`
	Thumbnail  string `json:"thumbnail"`
	VideoUrl   string `json:"video_url"`
	IsVertical bool   `json:"is_vertical"`
}

type ListAdCampaignsResponseItem struct {
	AdCampaignId   int64                              `json:"ad_campaign_id"`
	Content        ListAdCampaignsResponseItemContent `json:"content"`
	Views          int                                `json:"views"`
	Clicks         int                                `json:"clicks"`
	Status         database.AdCampaignStatus          `json:"status"`
	Budget         decimal.Decimal                    `json:"budget"`
	OriginalBudget decimal.Decimal                    `json:"original_budget"`
}

type HasAdCampaignsResponse struct {
	HasAdCampaign bool `json:"has_ad_campaign"`
}
