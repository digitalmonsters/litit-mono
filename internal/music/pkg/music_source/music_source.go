package music_source

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/music/configs"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/digitalmonsters/music/pkg/music_source/internal"
	"github.com/digitalmonsters/music/pkg/music_source/internal/lit"
	"github.com/digitalmonsters/music/pkg/music_source/internal/soundstripe"
	"github.com/pkg/errors"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"time"
)

type MusicStorageService struct {
	cfg             *configs.Settings
	implementations map[database.SongSource]internal.IMusicStorageAdapter
}

func NewMusicStorageService(configuration *configs.Settings) *MusicStorageService {
	return &MusicStorageService{
		cfg: configuration,
		implementations: map[database.SongSource]internal.IMusicStorageAdapter{
			database.SongSourceOwnStorage:  lit.NewService(),
			database.SongSourceSoundStripe: soundstripe.NewService(*configuration.SoundStripe),
		},
	}
}

func (s *MusicStorageService) findMusicInPlaylists(playlistIds []int64, source database.SongSource, page int, size int, db *gorm.DB, ctx context.Context) (*ListMusicResponse, error) {
	limit, offset := 0, 0
	limit = size
	offset = (page - 1) * size

	var dbSong []database.Song

	query := db.Model(dbSong).
		Joins("join playlist_song_relations psr on psr.song_id = songs.id").
		Joins("join playlists p on p.id = psr.playlist_id and p.deleted_at is null").
		Where("p.id in ? and songs.deleted_at is null", playlistIds)

	if source > 0 {
		query = query.Where("songs.source = ?", source)
	}

	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	if err := query.Limit(limit).Offset(offset).Find(&dbSong).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	var songs []internal.SongModel
	for _, song := range dbSong {
		songs = append(songs, internal.SongModel{
			Source:       song.Source,
			ExternalId:   song.ExternalId,
			Title:        song.Title,
			Artist:       song.Artist,
			ImageUrl:     song.ImageUrl,
			Genre:        song.Genre,
			Duration:     song.Duration,
			Files:        nil,
			DateUploaded: null.TimeFrom(song.CreatedAt),
			Playlists:    nil,
		})
	}

	songs, err := s.fillPlaylists(songs, source, db)
	if err != nil {
		apm_helper.LogError(err, ctx)
	}

	return &ListMusicResponse{
		Songs:      songs,
		TotalCount: totalCount,
	}, nil

}

func (s *MusicStorageService) ListMusic(req ListMusicRequest, db *gorm.DB, apmTx *apm.Transaction, ctx context.Context) (*ListMusicResponse, error) {
	impl, err := s.getImplementation(req.Source)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if len(req.PlaylistIds) > 0 {
		return s.findMusicInPlaylists(req.PlaylistIds, req.Source, req.Page, req.Size, db, ctx)
	}

	respCh := <-impl.GetSongsList(internal.GetSongsListRequest{
		SearchKeyword: req.SearchKeyword,
		Page:          req.Page,
		Size:          req.Size,
	}, db, apmTx, ctx)

	if respCh.Error != nil {
		return nil, errors.WithStack(respCh.Error)
	}

	if len(respCh.Response.Songs) > 0 {
		respCh.Response.Songs, err = s.fillPlaylists(respCh.Response.Songs, req.Source, db)
		if err != nil {
			apm_helper.LogError(err, ctx)
		}
	}

	return &ListMusicResponse{
		Songs:      respCh.Response.Songs,
		TotalCount: respCh.Response.TotalCount,
	}, nil
}

func (s *MusicStorageService) fillPlaylists(songs []internal.SongModel, source database.SongSource, db *gorm.DB) ([]internal.SongModel, error) {
	var externalIds []string
	for _, song := range songs {
		externalIds = append(externalIds, song.ExternalId)
	}

	var playlists []struct {
		Id           int64
		ExternalId   string
		PlaylistId   int64
		PlaylistName string
		CreatedAt    time.Time
	}

	if err := db.Model(&database.Song{}).
		Select("songs.external_id",
			"songs.id",
			"playlists.id as playlist_id",
			"playlists.name as playlist_name",
			"songs.created_at").
		Joins("left join playlist_song_relations psr on psr.song_id = songs.id").
		Joins("left join playlists on playlists.id = psr.playlist_id").
		Where("source = ? and external_id in ? and playlists.deleted_at is null", source, externalIds).
		Find(&playlists).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	for ind, song := range songs {
		for _, p := range playlists {
			if song.ExternalId == p.ExternalId {
				if p.PlaylistId > 0 {
					songs[ind].Playlists = append(songs[ind].Playlists, internal.PlaylistModel{
						Id:   p.PlaylistId,
						Name: p.PlaylistName,
					})
				}

				songs[ind].DateUploaded = null.TimeFrom(p.CreatedAt)
			}
		}
	}

	return songs, nil
}

func (s *MusicStorageService) SyncMusic(externalMusicIds []string, source database.SongSource, tx *gorm.DB, apmTransaction *apm.Transaction, ctx context.Context) error {
	impl, err := s.getImplementation(source)
	if err != nil {
		return errors.WithStack(err)
	}

	return impl.SyncSongsList(externalMusicIds, tx, apmTransaction, ctx)
}

func (s *MusicStorageService) GetMusicUrl(externalId string, source database.SongSource, db *gorm.DB, apmTx *apm.Transaction, ctx context.Context) (map[string]string, error) {
	impl, err := s.getImplementation(source)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	data, err := impl.GetSongUrl(externalId, db, apmTx, ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return data, nil
}

func (s *MusicStorageService) getImplementation(implType database.SongSource) (internal.IMusicStorageAdapter, error) {
	impl, ok := s.implementations[implType]
	if ok {
		return impl, nil
	}

	return nil, errors.New(fmt.Sprintf("muscic adapter [%v] not implemented", implType))
}
