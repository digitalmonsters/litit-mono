package common

import (
	"github.com/digitalmonsters/ads-manager/pkg/database"
	"github.com/shopspring/decimal"
	"gopkg.in/guregu/null.v4"
	"time"
)

type UpsertActionButtonsRequest struct {
	Items []UpsertButtonItem `json:"items"`
}

type UpsertButtonItem struct {
	Id   null.Int            `json:"id"`
	Name string              `json:"name"`
	Type database.ButtonType `json:"type"`
}

type UpsertRejectReasonsRequest struct {
	Items []UpsertRejectReason `json:"items"`
}

type UpsertRejectReason struct {
	Id     null.Int `json:"id"`
	Reason string   `json:"reason"`
}

type DeleteRequest struct {
	Ids []int64 `json:"ids"`
}

type ListActionButtonsRequest struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type PublicListActionButtonsRequest struct {
	Limit  int      `json:"limit"`
	Offset int      `json:"offset"`
	Type   null.Int `json:"type"`
}

type ListActionButtonsResponse struct {
	Items      []ActionButtonModel `json:"items"`
	TotalCount int64               `json:"total_count"`
}

type ActionButtonModel struct {
	Id   int64               `json:"id"`
	Type database.ButtonType `json:"type"`
	Name string              `json:"name"`
}

type ListRejectReasonsRequest struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type ListRejectReasonsResponse struct {
	Items      []RejectReasonModel `json:"items"`
	TotalCount int64               `json:"total_count"`
}

type RejectReasonModel struct {
	Id     int64  `json:"id"`
	Reason string `json:"reason"`
}

type AddModerationItem struct {
	Id             int64                     `json:"id"`
	UserId         int64                     `json:"user_id"`
	Username       string                    `json:"username"`
	FirstName      string                    `json:"first_name"`
	LastName       string                    `json:"last_name"`
	Email          string                    `json:"email"`
	Name           string                    `json:"name"`
	AdType         database.AdType           `json:"ad_type"`
	Status         database.AdCampaignStatus `json:"status"`
	ContentId      int64                     `json:"content_id"`
	Link           null.String               `json:"link"`
	Country        null.String               `json:"country"`
	CreatedAt      time.Time                 `json:"created_at"`
	StartedAt      null.Time                 `json:"started_at"`
	EndedAt        null.Time                 `json:"ended_at"`
	DurationMin    uint                      `json:"duration_min"`
	Budget         decimal.Decimal           `json:"budget"`
	OriginalBudget decimal.Decimal           `json:"original_budget"`
	Gender         null.String               `json:"gender"`
	AgeFrom        uint                      `json:"age_from"`
	AgeTo          uint                      `json:"age_to"`
	RejectReasonId null.Int                  `json:"reject_reason_id"`
	SlaExpired     bool                      `json:"sla_expired"`
	Thumbnail      string                    `json:"thumbnail"`
	VideoUrl       string                    `json:"video_url"`
	AnimUrl        string                    `json:"anim_url"`
}

type UpsertAdCampaignCountryPriceRequest struct {
	Items []AdCampaignCountryPriceItemModel `json:"items"`
}

type AdCampaignCountryPriceItemModel struct {
	CountryCode   string          `json:"country_code"`
	Price         decimal.Decimal `json:"price"`
	CountryName   string          `json:"country_name"`
	IsGlobalPrice bool            `json:"is_global_price"`
}

type ListAdCampaignCountryPriceRequest struct {
	CountryCode   null.String         `json:"country_code"`
	CountryName   null.String         `json:"country_name"`
	PriceFrom     decimal.NullDecimal `json:"price_from"`
	PriceTo       decimal.NullDecimal `json:"price_to"`
	IsGlobalPrice null.Bool           `json:"is_global_price"`
	Limit         int                 `json:"limit"`
	Offset        int                 `json:"offset"`
}

type ListAdCampaignCountryPriceResponse struct {
	Items      []AdCampaignCountryPriceItemModel `json:"items"`
	TotalCount int64                             `json:"total_count"`
}
