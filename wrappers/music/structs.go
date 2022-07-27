package music

import (
	"gopkg.in/guregu/null.v4"
	"time"
)

type GetMusicInternalRequests struct {
	Ids []int64 `json:"ids"`
}

type CreatorSongStatus int

const (
	CreatorSongStatusNone      = CreatorSongStatus(0)
	CreatorSongStatusPublished = CreatorSongStatus(1)
	CreatorSongStatusRejected  = CreatorSongStatus(2)
	CreatorSongStatusApproved  = CreatorSongStatus(3)
)

type SimpleMusic struct {
	Id                int64             `json:"id"`
	UserId            int64             `json:"user_id"`
	Name              string            `json:"name"`
	Status            CreatorSongStatus `json:"status"`
	LyricAuthor       null.String       `json:"lyric_author"`
	MusicAuthor       string            `json:"music_author"`
	CategoryId        int64             `json:"category_id"`
	MoodId            int64             `json:"mood_id"`
	FullSongUrl       string            `json:"full_song_url"`
	FullSongDuration  float64           `json:"full_song_duration"`
	ShortSongUrl      string            `json:"short_song_url"`
	ShortSongDuration float64           `json:"short_song_duration"`
	ImageUrl          string            `json:"image_url"`
	Hashtags          []string          `json:"hashtags"`
	ShortListens      int               `json:"short_listens"`
	FullListens       int               `json:"full_listens"`
	Likes             int               `json:"likes"`
	Comments          int               `json:"comments"`
	UsedInVideo       int               `json:"used_in_video"`
	CreatedAt         time.Time         `json:"created_at"`
}
