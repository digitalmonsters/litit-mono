package song

import (
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/digitalmonsters/music/pkg/music_source"
	"github.com/pkg/errors"
	"github.com/thoas/go-funk"
	"go.elastic.co/apm"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func AddSongToPlaylistBulk(req AddSongToPlaylistRequest, db *gorm.DB, apmTransaction *apm.Transaction, musicStorageService *music_source.MusicStorageService) error {
	tx := db.Begin()
	defer tx.Rollback()

	var externalSongIds []string
	for _, s := range req.Songs {
		if !funk.ContainsString(externalSongIds, s.ExternalSongId) {
			externalSongIds = append(externalSongIds, s.ExternalSongId)
		}
	}

	err := musicStorageService.SyncMusic(externalSongIds, req.Source, tx, apmTransaction)
	if err != nil {
		return errors.WithStack(err)
	}

	var songsRelation []struct {
		Id         int64
		ExternalId string
	}

	if err := tx.Model(&database.Song{}).
		Where("external_id in ? and source = ?", externalSongIds, req.Source).
		Find(&songsRelation).Error; err != nil {
		return errors.WithStack(err)
	}

	if len(externalSongIds) != len(songsRelation) {
		return errors.New("something went wrong (sync error)")
	}

	songsRelationsMapped := map[string]int64{}
	for _, r := range songsRelation {
		songsRelationsMapped[r.ExternalId] = r.Id
	}

	for _, s := range req.Songs {
		internalSongId, ok := songsRelationsMapped[s.ExternalSongId]
		if !ok {
			continue
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "playlist_id"}, {Name: "song_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"sort_order": s.SortOrder,
			}),
		}).Create(&database.PlaylistSongRelations{
			PlaylistId: s.PlaylistId,
			SongId:     internalSongId,
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
	tx := db.Begin()
	defer tx.Rollback()

	for _, item := range req.Items {
		if item.PlaylistId == 0 {
			continue
		}

		if len(item.SongsIds) == 0 {
			continue
		}

		if err := tx.Where("playlist_id = ? and song_id in ?", item.PlaylistId, item.SongsIds).Delete(&database.PlaylistSongRelations{}).Error; err != nil {
			return errors.WithStack(err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func PlaylistSongListAdmin(req PlaylistSongListRequest, db *gorm.DB) (*PlaylistSongListResponse, error) {
	if req.PlaylistId == 0 {
		return nil, errors.New("playlist_id is required")
	}

	var totalCount int64

	query := db.Table("playlist_song_relations psr").
		Joins("join songs on psr.song_id = songs.id").
		Where("psr.playlist_id = ?", req.PlaylistId)

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	var songs []database.Song

	if err := query.Order("sort_order desc").
		Limit(req.Limit).Offset(req.Offset).
		Find(&songs).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return &PlaylistSongListResponse{
		Items:      songs,
		TotalCount: totalCount,
	}, nil
}

func GetSongUrl(req GetSongUrlRequest, db *gorm.DB, apmTransaction *apm.Transaction, service *music_source.MusicStorageService) (map[string]string, error) {
	tx := db.Begin()
	defer tx.Rollback()

	var song database.Song
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&song, req.SongId).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	data, err := service.GetMusicUrl(song.ExternalId, song.Source, db, apmTransaction)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if err = tx.Model(&song).Where("id = ?", song.Id).
		Update("listen_amount", gorm.Expr("listen_amount + 1")).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	if err = tx.Commit().Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return data, nil
}
