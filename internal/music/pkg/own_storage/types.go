package own_storage

import (
	"github.com/digitalmonsters/music/pkg/database"
	"gopkg.in/guregu/null.v4"
)

type AddSongsToOwnStorageRequest struct {
	Items []OwnSongItem `json:"items"`
}

type OwnSongItem struct {
	Id          null.Int `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Artist      string   `json:"artist"`
	ImageUrl    string   `json:"image_url"`
	FileUrl     string   `json:"file_url"`
	Genre       string   `json:"genre"`
	Duration    float64  `json:"duration"`
}

type DeleteSongsFromOwnStorageRequest struct {
	SongIds []int64 `json:"song_ids"`
}

type OwnStorageMusicListRequest struct {
	SearchKeyword null.String `json:"search_keyword"`
	Order         OrderOption `json:"order"`
	Limit         int         `json:"limit"`
	Offset        int         `json:"offset"`
}

type OrderOption uint8

const (
	OrderNone            OrderOption = 0
	OrderDurationAsc     OrderOption = 1
	OrderDurationDesc    OrderOption = 2
	OrderDateCreatedAsc  OrderOption = 3
	OrderDateCreatedDesc OrderOption = 4
)

type OwnStorageMusicListResponse struct {
	Items      []database.MusicStorage `json:"items"`
	TotalCount int64                   `json:"total_count"`
}
