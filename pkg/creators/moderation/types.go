package moderation

import (
	"github.com/digitalmonsters/music/pkg/database"
	"gopkg.in/guregu/null.v4"
)

type RejectMusicRequest struct {
	SongId       int64 `json:"song_id"`
	RejectReason int64 `json:"reject_reason"`
}

type ApproveMusicRequest struct {
	SongId int64 `json:"song_id"`
}

type ListRequest struct {
	UserId     null.Int                     `json:"user_id"`
	Status     []database.CreatorSongStatus `json:"status"`
	Keyword    null.String                  `json:"keyword"`
	CategoryId null.Int                     `json:"category_id"`
	MoodId     null.Int                     `json:"mood_id"`
	Limit      int                          `json:"limit"`
	Offset     int                          `json:"offset"`
}

type ListResponse struct {
	Items      []listItem `json:"items"`
	TotalCount int64      `json:"total_count"`
}

type listItem struct {
	SongName          string                     `json:"song_name"`
	Status            database.CreatorSongStatus `json:"status"`
	LyricAuthor       null.String                `json:"lyric_author"`
	MusicAuthor       string                     `json:"music_author"`
	CategoryId        int64                      `json:"category_id"`
	MoodId            int64                      `json:"mood_id"`
	FullSongUrl       string                     `json:"full_song_url"`
	FullSongDuration  float64                    `json:"full_song_duration"`
	ShortSongUrl      string                     `json:"short_song_url"`
	ShortSongDuration float64                    `json:"short_song_duration"`
	ImageUrl          string                     `json:"image_url"`

	UserId   int64  `json:"user_id"`
	Username string `json:"username"`
	Category string `json:"category"`
	Mood     string `json:"mood"`
}
