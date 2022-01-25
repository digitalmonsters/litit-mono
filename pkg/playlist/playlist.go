package playlist

import (
	"fmt"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/digitalmonsters/music/pkg/frontend"
	"github.com/pilagod/gorm-cursor-paginator/v2/paginator"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

func UpsertPlaylist(req UpsertPlaylistRequest, db *gorm.DB) (*database.Playlist, error) {
	if len(req.Name) == 0 {
		return nil, errors.New("playlist name is empty")
	}

	tx := db.Begin()
	defer tx.Rollback()

	playlist := database.Playlist{
		Name:      req.Name,
		SortOrder: req.SortOrder,
		Color:     req.Color,
		IsActive:  req.IsActive,
		CreatedAt: time.Now(),
	}

	if req.Id.Valid {
		playlist.Id = req.Id.Int64
	}

	if err := tx.Model(&playlist).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			UpdateAll: true,
		}).
		Create(&playlist).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return &playlist, nil
}

func DeletePlaylistsBulk(req DeletePlaylistsBulkRequest, db *gorm.DB) error {
	if len(req.Ids) == 0 {
		return errors.New("nothing to delete")
	}

	tx := db.Begin()
	defer tx.Rollback()

	if err := tx.Delete(&database.Playlist{}, req.Ids).Error; err != nil {
		return errors.WithStack(err)
	}

	if err := tx.Commit().Error; err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func PlaylistListingAdmin(req PlaylistListingAdminRequest, db *gorm.DB) (*PlaylistListingAdminResponse, error) {
	var playlists []database.Playlist
	query := db.Model(&playlists)

	if req.Name.Valid {
		query = query.Where("name ilike ?", fmt.Sprintf("%%%v%%", req.Name.String))
	}

	if req.IsActive.Valid {
		query = query.Where("is_active = ?", req.IsActive.Bool)
	}

	if req.Order > 0 {
		switch req.Order {
		case OrderSortOrderAsc:
			query = query.Order("sort_order asc")
		case OrderSortOrderDesc:
			query = query.Order("sort_order desc")
		case OrderSongsCountAsc:
			query = query.Order("songs_count asc")
		case OrderSongsCountDesc:
			query = query.Order("songs_count desc")
		}
	}

	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	if err := query.Order("id desc").
		Limit(req.Limit).Offset(req.Offset).
		Find(&playlists).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return &PlaylistListingAdminResponse{
		Items:      playlists,
		TotalCount: totalCount,
	}, nil
}

func PlaylistListingPublic(req PlayListListingPublicRequest, db *gorm.DB) (*PlayListListingPublicResponse, error) {
	var playlists database.Playlists

	query := db.Model(&database.Playlist{}).Where("songs_count > 0 and is_active = true")

	if req.Name.Valid {
		query = query.Where("name ilike ?", fmt.Sprintf("%%%v%%", req.Name.String))
	}

	paginatorRules := []paginator.Rule{
		{
			Key:   "SortOrder",
			Order: paginator.DESC,
		},
		{
			Key:   "Id",
			Order: paginator.DESC,
		},
	}

	p := paginator.New(
		&paginator.Config{
			Rules: paginatorRules,
			Limit: req.Count,
		},
	)

	if len(req.Cursor) > 0 {
		p.SetAfterCursor(req.Cursor)
	}

	result, cursor, err := p.Paginate(query, &playlists)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	if result.Error != nil {
		return nil, errors.WithStack(result.Error)
	}

	resp := &PlayListListingPublicResponse{
		Items: make([]frontend.Playlist, 0),
	}

	if len(playlists) == 0 {
		return resp, nil
	}

	resp.Items = playlists.ConvertToFrontendModel()

	if cursor.After != nil {
		resp.Cursor = *cursor.After
	}

	return resp, nil
}

func PlaylistSongsListPublic(req PlaylistSongsListPublicRequest, db *gorm.DB) (*PlaylistSongsListPublicResponse, error) {
	if req.PlaylistId == 0 {
		return nil, errors.New("playlist is required")
	}

	var relations []database.PlaylistSongRelations

	query := db.Table("playlist_song_relations").Where("playlist_id = ?", req.PlaylistId).Debug()

	paginatorRules := []paginator.Rule{
		{
			Key:   "SortOrder",
			Order: paginator.DESC,
		},
		{
			Key:   "SongId",
			Order: paginator.DESC,
		},
	}

	p := paginator.New(
		&paginator.Config{
			Rules: paginatorRules,
			Limit: req.Count,
		},
	)

	if len(req.Cursor) > 0 {
		p.SetAfterCursor(req.Cursor)
	}

	result, cursor, err := p.Paginate(query, &relations)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	if result.Error != nil {
		return nil, errors.WithStack(result.Error)
	}

	var songIds []int64
	for _, sr := range relations {
		songIds = append(songIds, sr.SongId)
	}

	var dbSongs database.Songs
	if err := db.Find(&dbSongs, songIds).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	var songs database.Songs
	for _, songId := range songIds {
		for _, s := range dbSongs {
			if songId == s.Id {
				songs = append(songs, s)
			}
		}
	}

	resp := &PlaylistSongsListPublicResponse{
		Items: make([]frontend.Song, 0),
	}

	if len(songs) == 0 {
		return resp, nil
	}

	resp.Items = songs.ConvertToFrontendModel()

	if cursor.After != nil {
		resp.Cursor = *cursor.After
	}

	return resp, nil
}
