package playlist

import (
	"fmt"
	"github.com/digitalmonsters/music/pkg/database"
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
		CreatedAt: time.Now(),
	}

	if req.Id.Valid {
		playlist.Id = req.Id.Int64
	}

	if err := tx.Model(&playlist).
		Clauses(clause.OnConflict{UpdateAll: true}).
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

	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	if err := db.Order("id desc").
		Limit(req.Limit).Offset(req.Offset).
		Find(&playlists).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return &PlaylistListingAdminResponse{
		Playlists:  playlists,
		TotalCount: totalCount,
	}, nil
}
