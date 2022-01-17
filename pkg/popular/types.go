package popular

import "github.com/digitalmonsters/music/pkg/frontend"

type GetPopularSongsRequest struct {
	Count  int    `json:"count"`
	Cursor string `json:"cursor"`
}

type GetPopularSongsResponse struct {
	Songs  []frontend.Song `json:"songs"`
	Cursor string          `json:"cursor"`
}
