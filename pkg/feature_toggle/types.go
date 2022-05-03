package feature_toggle

import (
	"github.com/digitalmonsters/configurator/pkg/database"
	"gopkg.in/guregu/null.v4"
	"time"
)

type GetFeatureTogglesRequest struct {
	Keys        []string `json:"keys"`
	Limit       int      `json:"limit"`
	Offset      int      `json:"offset"`
	HideDeleted bool     `json:"show_deleted"`
}
type GetFeatureTogglesResponse struct {
	Items      []FeatureToggleModel `json:"items"`
	TotalCount int64                `json:"total_count"`
}
type FeatureToggleModel struct {
	Id        int64                        `json:"id"`
	Key       string                       `json:"key"`
	Value     database.FeatureToggleConfig `json:"value"`
	CreatedAt time.Time                    `json:"created_at"`
	UpdatedAt time.Time                    `json:"updated_at"`
	DeletedAt null.Time                    `json:"deleted_at"`
}
type CreateFeatureToggleRequest struct {
	Key   string                       `json:"key"`
	Value database.FeatureToggleConfig `json:"value"`
}
type UpdateFeatureToggleRequest struct {
	Id    int64                        `json:"id"`
	Value database.FeatureToggleConfig `json:"value"`
}
type DeleteFeatureToggleRequest struct {
	Id int64 `json:"id"`
}
type CreateFeatureToggleEventsRequest struct {
	Events []database.FeatureEvent `json:"events"`
}
type ListFeatureToggleEventsRequest struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}
type ListFeatureToggleEventsResponse struct {
	TotalCount int64                         `json:"total_count"`
	Items      []database.FeatureToggleEvent `json:"items"`
}
