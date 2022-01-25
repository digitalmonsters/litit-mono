package internal

import (
	"github.com/digitalmonsters/music/pkg/database"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
)

type IMusicStorageAdapter interface {
	SyncSongsList(songIds []string, tx *gorm.DB, apmTransaction *apm.Transaction) error
	GetSongsList(req GetSongsListRequest, db *gorm.DB, apmTransaction *apm.Transaction) chan GetSongsListResponseChan
	GetSongUrl(externalSongId string, db *gorm.DB, apmTransaction *apm.Transaction) (map[string]string, error)
}

type GetSongsListResponseChan struct {
	Error    error
	Response GetSongsListResponse `json:"response"`
}

type SongModel struct {
	Source       database.SongSource `json:"source"`
	ExternalId   string              `json:"external_id"`
	Title        string              `json:"title"`
	Artist       string              `json:"artist"`
	ImageUrl     string              `json:"image_url"`
	Genre        string              `json:"genre"`
	Duration     float64             `json:"duration"`
	Files        map[string]string   `json:"files"`
	DateUploaded null.Time           `json:"date_uploaded"`
	Playlists    []PlaylistModel     `json:"playlists"`
}

type PlaylistModel struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
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
