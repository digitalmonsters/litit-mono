package categories

import (
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/digitalmonsters/music/pkg/frontend"
	"gopkg.in/guregu/null.v4"
)

type UpsertRequest struct {
	Items []category `json:"items"`
}

type category struct {
	Id        null.Int `json:"id"`
	Name      string   `json:"name"`
	IsActive  bool     `json:"is_active"`
	SortOrder int      `json:"sort_order"`
}

type ListRequest struct {
	Name       null.String `json:"name"`
	Limit      int         `json:"limit"`
	Offset     int         `json:"offset"`
	SortOption SortOption  `json:"sort_option"`
}

type SortOption int

const (
	SortOptionNone           = SortOption(0)
	SortOptionSortOrderDesc  = SortOption(1)
	SortOptionSortOrderAsc   = SortOption(2)
	SortOptionSongsCountDesc = SortOption(3)
	SortOptionSongsCountAsc  = SortOption(4)
)

type ListResponse struct {
	Items      []database.Category `json:"items"`
	TotalCount int64               `json:"total_count"`
}

type DeleteRequest struct {
	Ids []int64 `json:"ids"`
}

type PublicListRequest struct {
	Name   null.String `json:"name"`
	Count  int         `json:"count"`
	Cursor string      `json:"cursor"`
}

type PublicListResponse struct {
	Items  []frontend.Category `json:"items"`
	Cursor string              `json:"cursor"`
}
