package song

import (
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func AddSongToPlaylistBulk(req AddSongToPlaylistRequest, db *gorm.DB) error {
	tx := db.Begin()
	defer tx.Rollback()

	//todo: soundstripe song validation logic

	for _, s := range req.Songs {
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "playlist_id"}, {Name: "song_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"sort_order": s.SortOrder,
			}),
		}).Create(&database.PlaylistSongRelations{
			PlaylistId: s.PlaylistId,
			SongId:     s.SongId,
			SortOrder:  s.SortOrder,
		}).Error; err != nil {
			return errors.WithStack(err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func DeleteSongFromPlaylistsBulk(req DeleteSongsFromPlaylistBulkRequest, db *gorm.DB) error {
	if req.PlaylistId == 0 {
		return errors.New("playlist is required")
	}

	if len(req.SongsIds) == 0 {
		return errors.New("songs_ids is required")
	}

	tx := db.Begin()
	defer tx.Rollback()

	if err := tx.Where("playlist_id = ? and song_id in ?", req.PlaylistId, req.SongsIds).Debug().Delete(&database.PlaylistSongRelations{}).Error; err != nil {
		return errors.WithStack(err)
	}

	if err := tx.Commit().Error; err != nil {
		return errors.WithStack(err)
	}

	return nil
}
