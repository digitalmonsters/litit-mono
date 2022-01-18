package popular

import (
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/pilagod/gorm-cursor-paginator/v2/paginator"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func GetAudioUrl(songId string, db *gorm.DB) (interface{}, error) {
	//todo: audio url parse logic

	if err := db.Model(&database.Song{}).Where("id = ?", songId).
		Update("listen_amount", gorm.Expr("listen_amount + 1")).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return nil, nil
}

func GetPopularSongs(req GetPopularSongsRequest, db *gorm.DB) (*GetPopularSongsResponse, error) {
	var songs database.Songs

	query := db.Table("songs").Where("listen_amount  > 0")

	paginatorRules := []paginator.Rule{
		{
			Key:   "ListenAmount",
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

	result, cursor, err := p.Paginate(query, &songs)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	if result.Error != nil {
		return nil, errors.WithStack(result.Error)
	}

	resp := &GetPopularSongsResponse{
		Items: songs.ConvertToFrontendModel(),
	}

	if cursor.After != nil {
		resp.Cursor = *cursor.After
	}

	return resp, nil
}
