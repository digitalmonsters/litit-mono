package music_source

import (
	"fmt"
	"github.com/digitalmonsters/music/configs"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/digitalmonsters/music/pkg/music_source/internal"
	"github.com/digitalmonsters/music/pkg/music_source/internal/soundstripe"
	"github.com/pkg/errors"
	"go.elastic.co/apm"
	"gorm.io/gorm"
)

type MusicStorageService struct {
	cfg *configs.Settings
}

func NewMusicStorageService(configuration *configs.Settings) *MusicStorageService {
	return &MusicStorageService{
		cfg: configuration,
	}
}

func (s *MusicStorageService) ListMusic(req ListMusicRequest, apmTransaction *apm.Transaction) (*ListMusicResponse, error) {
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

	return &ListMusicResponse{
		Songs:      respCh.Response.Songs,
		TotalCount: respCh.Response.TotalCount,
	}, nil
}

func (s *MusicStorageService) SyncMusic(externalMusicIds []string, source database.SongSource, db *gorm.DB, apmTransaction *apm.Transaction) error {
	impl, err := s.getImplementation(source)
	if err != nil {
		return errors.WithStack(err)
	}

	return impl.SyncSongsList(externalMusicIds, db, apmTransaction)
}

func (s *MusicStorageService) getImplementation(implType database.SongSource) (internal.IMusicStorageAdapter, error) {
	switch implType {
	case database.SongSourceSoundStripe:
		return soundstripe.NewService(*s.cfg.SoundStripe), nil
	default:
		return nil, errors.New(fmt.Sprintf("muscic adapter [%v] not implemented", implType))
	}
}
