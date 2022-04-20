package reject_reasons

import (
	"github.com/digitalmonsters/music/pkg/database"
	"gopkg.in/guregu/null.v4"
)

type UpsertRequest struct {
	Items []rejectReason `json:"items"`
}

type rejectReason struct {
	Id     null.Int            `json:"id"`
	Type   database.ReasonType `json:"type"`
	Reason string              `json:"reason"`
}

type ListRequest struct {
	Type   database.ReasonType `json:"type"`
	Limit  int                 `json:"limit"`
	Offset int                 `json:"offset"`
}

type ListResponse struct {
	Items      []database.CreatorRejectReasons `json:"items"`
	TotalCount int64                           `json:"total_count"`
}

type DeleteRequest struct {
	Ids []int64 `json:"ids"`
}
