package popular

import (
	"fmt"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/digitalmonsters/music/pkg/frontend"
	"github.com/pilagod/gorm-cursor-paginator/v2/paginator"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func GetPopularSongs(req GetPopularSongsRequest, db *gorm.DB, executionData router.MethodExecutionData) (*GetPopularSongsResponse, error) {
	var songs []database.Song

	query := db.Table("songs").Debug()

	if req.SearchKeyword.Valid {
		query = query.Joins("join playlist_song_relations psr on psr.song_id = songs.id").
			Joins("join playlists on playlists.id = psr.playlist_id and playlists.deleted_at is null").
			Where("title ilike ?", fmt.Sprintf("%%%v%%", req.SearchKeyword.String)).
			Or("artist ilike ?", fmt.Sprintf("%%%v%%", req.SearchKeyword.String))
	}

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
		Items: make([]frontend.Song, 0),
	}

	if len(songs) == 0 {
		return resp, nil
	}

	resp.Items = frontend.ConvertSongsToFrontendModel(songs, executionData.UserId, db, executionData.ApmTransaction)

	if cursor.After != nil {
		resp.Cursor = *cursor.After
	}

	return resp, nil
}
