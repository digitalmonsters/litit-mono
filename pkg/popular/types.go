package popular

import "github.com/digitalmonsters/music/pkg/frontend"

type GetPopularSongsRequest struct {
	Count  int    `json:"count"`
	Cursor string `json:"cursor"`
}

type GetPopularSongsResponse struct {
	Items  []frontend.Song `json:"items"`
	Cursor string          `json:"cursor"`
}
