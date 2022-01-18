package favorites

import (
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/pilagod/gorm-cursor-paginator/v2/paginator"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func AddToFavorites(req AddToFavoritesRequest, db *gorm.DB) error {
	favorite := database.Favorite{
		UserId: req.UserId,
		SongId: req.SongId,
	}

	if err := db.Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(&favorite).Error; err != nil {
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

func FavoriteSongsList(req FavoriteSongsListRequest, userId int64, db *gorm.DB) (*FavoriteSongsListResponse, error) {
	var favorites []database.Favorite

	query := db.Model(&database.Favorite{}).Where("user_id = ?", userId).Debug()

	paginatorRules := []paginator.Rule{
		{
			Key:   "CreatedAt",
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

	result, cursor, err := p.Paginate(query, &favorites)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	if result.Error != nil {
		return nil, errors.WithStack(result.Error)
	}

	var songIds []int64
	for _, f := range favorites {
		songIds = append(songIds, f.SongId)
	}

	var dbSongs database.Songs
	if err := db.Find(&dbSongs, songIds).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	var songs database.Songs
	for _, songId := range songIds {
		for _, s := range dbSongs {
			if s.Id == songId {
				songs = append(songs, s)
			}
		}
	}

	resp := &FavoriteSongsListResponse{
		Items: songs.ConvertToFrontendModel(),
	}

	if cursor.After != nil {
		resp.Cursor = *cursor.After
	}

	return resp, nil
}
