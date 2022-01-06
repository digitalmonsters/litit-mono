package favorites

import (
	"github.com/digitalmonsters/music/pkg/database"
	"gorm.io/gorm"
)

func AddToFavorites(req AddToFavoritesRequest, db *gorm.DB) error {
	favorite := database.Favorite{
		UserId: req.UserId,
		SongId: req.SongId,
	}

	if err := db.Create(&favorite).Error; err != nil {
		return err
	}

	return nil
}

func RemoveFromFavorites(req RemoveFromFavoritesRequest, db *gorm.DB) error {
	favorite := database.Favorite{
		UserId: req.UserId,
		SongId: req.SongId,
	}

	if err := db.Delete(&favorite, "user_id = ? and song_id = ?", req.UserId, req.SongId).Error; err != nil {
		return err
	}

	return nil
}
