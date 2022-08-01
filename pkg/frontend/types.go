package frontend

import (
	"github.com/digitalmonsters/go-common/frontend"
	"github.com/digitalmonsters/go-common/wrappers/music"
	"github.com/lib/pq"
	"gopkg.in/guregu/null.v4"
	"time"
)

type Playlist struct {
	Id         int64  `json:"id"`
	Name       string `json:"name"`
	Color      string `json:"color"`
	SongsCount int    `json:"songs_count"`
}

type Song struct {
	Id         int64   `json:"id"`
	Title      string  `json:"title"`
	Artist     string  `json:"artist"`
	Url        string  `json:"url"`
	IsFavorite bool    `json:"is_favorite"`
	ImageUrl   string  `json:"image_url"`
	Genre      string  `json:"genre"`
	Duration   float64 `json:"duration"`
}

type Category struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type Mood struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type RejectReason struct {
	Id     int64  `json:"id"`
	Reason string `json:"reason"`
}

type CreatorSongModel struct {
	Id                int64                   `json:"id"`
	UserId            int64                   `json:"user_id"`
	Name              string                  `json:"name"`
	Status            music.CreatorSongStatus `json:"status"`
	LyricAuthor       null.String             `json:"lyric_author"`
	MusicAuthor       string                  `json:"music_author"`
	CategoryId        int64                   `json:"category_id"`
	MoodId            int64                   `json:"mood_id"`
	FullSongUrl       string                  `json:"full_song_url"`
	FullSongDuration  float64                 `json:"full_song_duration"`
	ShortSongUrl      string                  `json:"short_song_url"`
	ShortSongDuration float64                 `json:"short_song_duration"`
	ImageUrl          string                  `json:"image_url"`
	Hashtags          pq.StringArray          `gorm:"type:text[]" json:"hashtags"`
	ShortListens      int                     `json:"short_listens"`
	FullListens       int                     `json:"full_listens"`
	Likes             int                     `json:"likes"`
	Comments          int                     `json:"comments"`
	UsedInVideo       int                     `json:"used_in_video"`
	CreatedAt         time.Time               `json:"created_at"`

	IsCreatorFollowing bool `json:"is_creator_following"`
	IsFollowing        bool `json:"is_following"`

	User         frontend.VideoUserModel `json:"user"`
	Category     *Category               `json:"category"`
	Mood         *Mood                   `json:"mood"`
	RejectReason *RejectReason           `json:"reject_reason"`
}
