package song

import "github.com/digitalmonsters/music/pkg/database"

type AddSongToPlaylistRequest struct {
	Source database.SongSource `json:"source"`
	Songs  []relationItem      `json:"songs"`
}

type relationItem struct {
	ExternalSongId string `json:"external_song_id"`
	PlaylistId     int64  `json:"playlist_id"`
	SortOrder      int    `json:"sort_order"`
}

type DeleteSongsFromPlaylistBulkRequest struct {
	Items []itemForDeletion `json:"items"`
}

type itemForDeletion struct {
	PlaylistId int64   `json:"playlist_id"`
	SongsIds   []int64 `json:"songs_ids"`
}

type PlaylistSongListRequest struct {
	PlaylistId int64 `json:"playlist_id"`
	Limit      int   `json:"limit"`
	Offset     int   `json:"offset"`
}

type PlaylistSongListResponse struct {
	Items      []database.Song `json:"items"`
	TotalCount int64           `json:"total_count"`
}

type GetSongUrlRequest struct {
	SongId int64 `json:"song_id"`
}
