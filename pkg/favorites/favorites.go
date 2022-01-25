package favorites

import (
	"fmt"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/digitalmonsters/music/pkg/frontend"
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

func FavoriteSongsList(req FavoriteSongsListRequest, db *gorm.DB, executionData router.MethodExecutionData) (*FavoriteSongsListResponse, error) {
	var favorites []database.Favorite

	query := db.Model(&database.Favorite{})

	if req.SearchKeyword.Valid {
		query = query.Joins("join songs s on s.id = favorites.song_id").
			Where(db.Where("s.title ilike ?", fmt.Sprintf("%%%v%%", req.SearchKeyword.String)).
				Or("s.artist ilike ?", fmt.Sprintf("%%%v%%", req.SearchKeyword.String)))

	}

	query = query.Where("user_id = ?", executionData.UserId)

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

	resp := &FavoriteSongsListResponse{
		Items: make([]frontend.Song, 0),
	}

	if len(favorites) == 0 {
		return resp, nil
	}

	var songIds []int64
	for _, f := range favorites {
		songIds = append(songIds, f.SongId)
	}

	var dbSongs []database.Song
	if err = db.Find(&dbSongs, songIds).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	if len(dbSongs) == 0 {
		return resp, nil
	}

	resp.Items = frontend.ConvertSongsToFrontendModel(dbSongs, executionData.UserId, db, executionData.ApmTransaction)

	if cursor.After != nil {
		resp.Cursor = *cursor.After
	}

	return resp, nil
}
