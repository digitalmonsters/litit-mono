package own_storage

import (
	"fmt"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

func UpsertSongsToOwnStorageBulk(req AddSongsToOwnStorageRequest, db *gorm.DB) ([]database.MusicStorage, error) {
	tx := db.Begin()
	defer tx.Rollback()

	var songsToAdd []database.MusicStorage
	for _, song := range req.Items {
		s := database.MusicStorage{
			Title:       song.Title,
			Description: song.Description,
			Artist:      song.Artist,
			ImageUrl:    song.ImageUrl,
			Genre:       song.Genre,
			Duration:    song.Duration,
			Url:         song.FileUrl,
		}

		if song.Id.Valid {
			s.Id = song.Id.Int64
			s.UpdatedAt = time.Now()
		}

		songsToAdd = append(songsToAdd, s)
	}

	if err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(&songsToAdd).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return songsToAdd, nil
}

func DeleteSongsFromOwnStorageBulk(req DeleteSongsFromOwnStorageRequest, db *gorm.DB) error {
	tx := db.Begin()
	defer tx.Rollback()

	if err := tx.Delete(&database.MusicStorage{}, req.SongIds).Error; err != nil {
		return errors.WithStack(err)
	}

	return tx.Commit().Error
}

func OwnStorageMusicList(req OwnStorageMusicListRequest, db *gorm.DB) (*OwnStorageMusicListResponse, error) {
	var songs []database.MusicStorage

	query := db.Model(&songs)
	if req.SearchKeyword.Valid {
		query = query.Where("title ilike ?", fmt.Sprintf("%%%v%%", req.SearchKeyword.String)).
			Or("description ilike ?", fmt.Sprintf("%%%v%%", req.SearchKeyword.String)).
			Or("artist ilike ?", fmt.Sprintf("%%%v%%", req.SearchKeyword.String))
	}

	if req.Order > 0 {
		switch req.Order {
		case OrderDurationAsc:
			query = query.Order("duration asc")
		case OrderDurationDesc:
			query = query.Order("duration desc")
		case OrderDateCreatedAsc:
			query = query.Order("created_at asc")
		case OrderDateCreatedDesc:
			query = query.Order("created_at desc")
		}
	}

	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	if err := query.Limit(req.Limit).Offset(req.Offset).Order("id desc").Find(&songs).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return &OwnStorageMusicListResponse{
		Items:      songs,
		TotalCount: totalCount,
	}, nil
}
