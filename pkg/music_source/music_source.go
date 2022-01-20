package music_source

import (
	"fmt"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/music/configs"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/digitalmonsters/music/pkg/music_source/internal"
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
			database.SongSourceSoundStripe: soundstripe.NewService(*configuration.SoundStripe),
		},
	}
}

func (s *MusicStorageService) ListMusic(req ListMusicRequest, db *gorm.DB, apmTransaction *apm.Transaction) (*ListMusicResponse, error) {
	impl, err := s.getImplementation(req.Source)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	respCh := <-impl.GetSongsList(internal.GetSongsListRequest{
		SearchKeyword: req.SearchKeyword,
		Page:          req.Page,
		Size:          req.Size,
	}, apmTransaction)

	if respCh.Error != nil {
		return nil, errors.WithStack(respCh.Error)
	}

	if len(respCh.Response.Songs) > 0 {
		respCh.Response.Songs, err = s.fillPlaylists(respCh.Response.Songs, req.Source, db)
		if err != nil {
			apm_helper.CaptureApmError(err, apmTransaction)
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
		ExternalId   string
		PlaylistId   int64
		PlaylistName string
		CreatedAt    time.Time
	}

	if err := db.Model(&database.Song{}).
		Select("songs.external_id",
			"playlists.id as playlist_id",
			"playlists.name as playlist_name",
			"songs.created_at").
		Joins("left join playlist_song_relations psr on psr.song_id = songs.id").
		Joins("left join playlists on playlists.id = psr.playlist_id and playlists.deleted_at is null").
		Where("source = ? and external_id in ?", source, externalIds).
		Find(&playlists).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	for ind, song := range songs {
		for _, p := range playlists {
			if song.ExternalId == p.ExternalId {
				songs[ind].Playlists = append(songs[ind].Playlists, internal.PlaylistModel{
					Id:   p.PlaylistId,
					Name: p.PlaylistName,
				})

				songs[ind].DateUploaded = null.TimeFrom(p.CreatedAt)
			}
		}
	}

	return songs, nil
}

func (s *MusicStorageService) SyncMusic(externalMusicIds []string, source database.SongSource, tx *gorm.DB, apmTransaction *apm.Transaction) error {
	impl, err := s.getImplementation(source)
	if err != nil {
		return errors.WithStack(err)
	}

	return impl.SyncSongsList(externalMusicIds, tx, apmTransaction)
}

func (s *MusicStorageService) getImplementation(implType database.SongSource) (internal.IMusicStorageAdapter, error) {

	switch implType {
	case database.SongSourceSoundStripe:
		return s.implementations[database.SongSourceSoundStripe], nil
	default:
		return nil, errors.New(fmt.Sprintf("muscic adapter [%v] not implemented", implType))
	}
}
