package common

import (
	"github.com/digitalmonsters/ads-manager/pkg/database"
	"gopkg.in/guregu/null.v4"
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
