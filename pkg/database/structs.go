package database

import (
	"gorm.io/gorm"
	"time"
)

type Playlist struct {
	Id         int64          `json:"id" gorm:"primaryKey"`
	Name       string         `json:"name"`
	SortOrder  int            `json:"sort_order"`
	Color      string         `json:"color"`
	SongsCount int            `json:"songs_count"`
	CreatedAt  time.Time      `json:"created_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at"`
}

func (Playlist) TableName() string {
	return "playlists"
}

type Song struct {
	Id        string    `json:"id" gorm:"primaryKey"`
	Title     string    `json:"title"`
	Artist    string    `json:"artist"`
	Url       string    `json:"url"`
	ImageUrl  string    `json:"image_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Song) TableName() string {
	return "songs"
}

type PlaylistSongRelations struct {
	PlaylistId int64
	SongId     string
	SortOrder  int
}

func (Song) PlaylistSongRelations() string {
	return "playlist_song_relations"
}
