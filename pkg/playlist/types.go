package playlist

import (
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/digitalmonsters/music/pkg/frontend"
	"gopkg.in/guregu/null.v4"
)

type UpsertPlaylistRequest struct {
	Id        null.Int `json:"id"`
	Name      string   `json:"name"`
	SortOrder int      `json:"sort_order"`
	Color     string   `json:"color"`
}

type DeletePlaylistsBulkRequest struct {
	Ids []int64 `json:"ids"`
}

type PlaylistListingAdminRequest struct {
	Name   null.String `json:"name"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
}

type PlaylistListingAdminResponse struct {
	Playlists  []database.Playlist `json:"playlists"`
	TotalCount int64               `json:"total_count"`
}

type PlayListListingPublicRequest struct {
	Name   null.String `json:"name"`
	Count  int         `json:"count"`
	Cursor string      `json:"cursor"`
}

type PlayListListingPublicResponse struct {
	Playlists []frontend.Playlist `json:"playlists"`
	Cursor    string              `json:"cursor"`
}

type PlaylistSongsListPublicRequest struct {
	PlaylistId int64  `json:"playlist_id"`
	Count      int    `json:"count"`
	Cursor     string `json:"cursor"`
}

type PlaylistSongsListPublicResponse struct {
	Songs  []frontend.Song `json:"songs"`
	Cursor string          `json:"cursor"`
}
