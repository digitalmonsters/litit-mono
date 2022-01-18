package internal

import (
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
)

type IMusicStorageAdapter interface {
	SyncSongsList(songIds []string, db *gorm.DB, apmTransaction *apm.Transaction) error
	GetSongsList(req GetSongsListRequest, apmTransaction *apm.Transaction) chan GetSongsListResponseChan
}

type GetSongsListResponseChan struct {
	Error    error
	Response GetSongsListResponse `json:"response"`
}

type SongModel struct {
	ExternalId string  `json:"external_id"`
	Title      string  `json:"title"`
	Artist     string  `json:"artist"`
	Url        string  `json:"url"`
	ImageUrl   string  `json:"image_url"`
	Genre      string  `json:"genre"`
	Duration   float64 `json:"duration"`
}

type GetSongResponseChan struct {
	Error error
	Song  SongModel `json:"song"`
}

type GetSongsListResponse struct {
	Songs      []SongModel `json:"songs"`
	TotalCount int64       `json:"total_count"`
}

type GetSongsListRequest struct {
	SearchKeyword null.String `json:"search_keyword"`
	Page          int         `json:"page"`
	Size          int         `json:"size"`
}
