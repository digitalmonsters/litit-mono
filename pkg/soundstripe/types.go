package soundstripe

import (
	"gopkg.in/guregu/null.v4"
)

type GetSongsListRequest struct {
	SearchKeyword null.String `json:"search_keyword"`
	Page          int         `json:"page"`
	Size          int         `json:"size"`
}

type SongModel struct {
	Id       string  `json:"id" gorm:"primaryKey"`
	Title    string  `json:"title"`
	Artist   string  `json:"artist"`
	Url      string  `json:"url"`
	ImageUrl string  `json:"image_url"`
	Genre    string  `json:"genre"`
	Duration float64 `json:"duration"`
}

type GetSongsListResponse struct {
	Songs      []SongModel `json:"songs"`
	TotalCount int64       `json:"total_count"`
}

type soundstripeSongsResp struct {
	Links struct {
		Meta struct {
			TotalCount int64 `json:"total_count"`
		} `json:"meta"`
	} `json:"links"`
}

type GetSongsListResponseChan struct {
	Error    error
	Response GetSongsListResponse `json:"response"`
}

type GetSongResponseChan struct {
	Error error
	Song  SongModel `json:"song"`
}
