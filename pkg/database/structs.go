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
	IsActive   bool           `json:"is_active"`
	CreatedAt  time.Time      `json:"created_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at"`
}

func (Playlist) TableName() string {
	return "playlists"
}

type Song struct {
	Id           int64          `json:"id" gorm:"primaryKey"`
	Source       SongSource     `json:"source"`
	ExternalId   string         `json:"external_id"`
	Title        string         `json:"title"`
	Artist       string         `json:"artist"`
	ImageUrl     string         `json:"image_url"`
	Genre        string         `json:"genre"`
	Duration     float64        `json:"duration"`
	ListenAmount int            `json:"listen_amount"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at"`
}

type SongSource int

const (
	SongSourceOwnStorage  = SongSource(1)
	SongSourceSoundStripe = SongSource(2)
)

func (Song) TableName() string {
	return "songs"
}

type PlaylistSongRelations struct {
	PlaylistId int64
	SongId     int64
	SortOrder  int
}

type Favorite struct {
	UserId    int64
	SongId    int64
	CreatedAt time.Time
}

type MusicStorage struct {
	Id          int64          `json:"id" gorm:"primaryKey"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Artist      string         `json:"artist"`
	ImageUrl    string         `json:"image_url"`
	Genre       string         `json:"genre"`
	Duration    float64        `json:"duration"`
	Url         string         `json:"url"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at"`
}

func (MusicStorage) TableName() string {
	return "music_storage"
}
