package favorites

import (
	"github.com/digitalmonsters/music/pkg/frontend"
	"gopkg.in/guregu/null.v4"
)

type AddToFavoritesRequest struct {
	UserId int64 `json:"user_id"`
	SongId int64 `json:"song_id"`
}

type RemoveFromFavoritesRequest struct {
	UserId int64 `json:"user_id"`
	SongId int64 `json:"song_id"`
}

type FavoriteSongsListRequest struct {
	SearchKeyword null.String `json:"search_keyword"`
	Count         int         `json:"count"`
	Cursor        string      `json:"cursor"`
}

type FavoriteSongsListResponse struct {
	Items  []frontend.Song `json:"items"`
	Cursor string          `json:"cursor"`
}
