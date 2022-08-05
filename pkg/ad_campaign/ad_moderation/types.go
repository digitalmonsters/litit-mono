package ad_moderation

import (
	"github.com/digitalmonsters/ads-manager/pkg/common"
	"github.com/digitalmonsters/ads-manager/pkg/database"
	"gopkg.in/guregu/null.v4"
)

type GetAdModerationRequest struct {
	UserId               null.Int                    `json:"user_id"`
	Status               []database.AdCampaignStatus `json:"status"`
	StartedAtFrom        null.Time                   `json:"started_at_from"`
	StartedAtTo          null.Time                   `json:"started_at_to"`
	EndedAtFrom          null.Time                   `json:"ended_at_from"`
	EndedAtTo            null.Time                   `json:"ended_at_to"`
	MaxThresholdExceeded null.Bool                   `json:"max_threshold_exceeded"`
	Limit                int                         `json:"limit"`
	Offset               int                         `json:"offset"`
}
type GetAdModerationResponse struct {
	TotalCount int64                      `json:"total_count"`
	Items      []common.AddModerationItem `json:"items"`
}

type SetAdRejectReasonRequest struct {
	Id             int64    `json:"id"`
	RejectReasonId null.Int `json:"reject_reason_id"`
}
