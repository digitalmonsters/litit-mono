package song

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
