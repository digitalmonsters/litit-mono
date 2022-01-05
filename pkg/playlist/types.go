package playlist

import (
	"github.com/digitalmonsters/music/pkg/database"
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
