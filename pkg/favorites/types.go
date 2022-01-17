package favorites

import "github.com/digitalmonsters/music/pkg/frontend"

type AddToFavoritesRequest struct {
	UserId int64
	SongId string
}

type RemoveFromFavoritesRequest struct {
	UserId int64
	SongId string
}

type FavoriteSongsListRequest struct {
	Count  int    `json:"count"`
	Cursor string `json:"cursor"`
}

type FavoriteSongsListResponse struct {
	Songs  []frontend.Song `json:"songs"`
	Cursor string          `json:"cursor"`
}
