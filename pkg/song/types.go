package song

import "github.com/digitalmonsters/music/pkg/database"

type AddSongToPlaylistRequest struct {
	Songs []RelationItem `json:"songs"`
}

type RelationItem struct {
	SongId     string `json:"song_id"`
	PlaylistId int64  `json:"playlist_id"`
	SortOrder  int    `json:"sort_order"`
}

type DeleteSongsFromPlaylistBulkRequest struct {
	PlaylistId int64    `json:"playlist_id"`
	SongsIds   []string `json:"songs_ids"`
}

type PlaylistSongListRequest struct {
	PlaylistId int64 `json:"playlist_id"`
	Limit      int   `json:"limit"`
	Offset     int   `json:"offset"`
}

type PlaylistSongListResponse struct {
	Songs      []database.Song `json:"songs"`
	TotalCount int64           `json:"total_count"`
}
