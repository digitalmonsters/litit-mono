package own_storage

import (
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"time"
)

func UpsertSongsToOwnStorageBulk(req AddSongsToOwnStorageRequest, db *gorm.DB) ([]database.MusicStorage, error) {
	tx := db.Begin()
	defer tx.Rollback()

	var songsToAdd []database.MusicStorage
	for _, song := range req.Items {
		songsToAdd = append(songsToAdd, database.MusicStorage{
			Title:       song.Title,
			Description: song.Description,
			Artist:      song.Artist,
			ImageUrl:    song.ImageUrl,
			Genre:       song.Genre,
			Duration:    song.Duration,
			Url:         song.FileUrl,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		})
	}

	if err := tx.Create(&songsToAdd).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return songsToAdd, nil
}
