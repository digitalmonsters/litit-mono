package popular

import (
	"github.com/digitalmonsters/music/pkg/frontend"
	"gopkg.in/guregu/null.v4"
)

type GetPopularSongsRequest struct {
	SearchKeyword null.String `json:"search_keyword"`
	Count         int         `json:"count"`
	Cursor        string      `json:"cursor"`
}

type GetPopularSongsResponse struct {
	Items  []frontend.Song `json:"items"`
	Cursor string          `json:"cursor"`
}
