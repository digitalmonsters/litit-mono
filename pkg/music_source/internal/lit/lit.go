package lit

import (
	"fmt"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/digitalmonsters/music/pkg/music_source/internal"
	"github.com/digitalmonsters/music/pkg/own_storage"
	"github.com/pkg/errors"
	"github.com/thoas/go-funk"
	"go.elastic.co/apm"
	"gorm.io/gorm"
	"strings"
)

type Service struct{}

func NewService() internal.IMusicStorageAdapter {
	return &Service{}
}

func (s *Service) SyncSongsList(externalSongsIds []string, tx *gorm.DB, apmTransaction *apm.Transaction) error {
	var missing []string

	var songsInDb []string
	if err := tx.Model(database.Song{}).Where("external_id in ? and source = ?", externalSongsIds, database.SongSourceOwnStorage).
		Pluck("external_id", &songsInDb).Error; err != nil {
		return errors.WithStack(err)
	}

	for _, songId := range externalSongsIds {
		if !funk.ContainsString(songsInDb, songId) {
			missing = append(missing, songId)
		}
	}

	if len(missing) == 0 {
		return nil
	}

	var songsToAdd []database.Song

	var songs []database.MusicStorage
	if err := tx.Model(&songs).Where("id in ?", missing).Find(&songs).Error; err != nil {
		return errors.WithStack(err)
	}

	for _, song := range songs {
		songsToAdd = append(songsToAdd, database.Song{
			Source:     database.SongSourceOwnStorage,
			ExternalId: fmt.Sprint(song.Id),
			Title:      song.Title,
			Artist:     song.Artist,
			ImageUrl:   song.ImageUrl,
			Genre:      song.Genre,
			Duration:   song.Duration,
		})
	}

	if len(songsToAdd) > 0 {
		if err := tx.Create(&songsToAdd).Error; err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (s *Service) GetSongUrl(externalSongId string, db *gorm.DB, apmTransaction *apm.Transaction) (map[string]string, error) {
	var song database.MusicStorage
	if err := db.Model(&song).Where("id = ?", externalSongId).Find(&song).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	if song.Id == 0 {
		return nil, errors.New("song not found")
	}

	ext := "mp3"
	spl := strings.Split(song.Url, ".")
	if len(spl) >= 2 {
		ext = spl[len(spl)-1]
	}

	return map[string]string{
		ext: song.Url,
	}, nil

}

func (s *Service) GetSongsList(req internal.GetSongsListRequest, db *gorm.DB, apmTransaction *apm.Transaction) chan internal.GetSongsListResponseChan {
	ch := make(chan internal.GetSongsListResponseChan, 2)

	go func() {
		finalResponse := internal.GetSongsListResponseChan{
			Response: internal.GetSongsListResponse{},
		}

		limit, offset := 0, 0
		limit = req.Size
		offset = (req.Page - 1) * req.Size

		resp, err := own_storage.OwnStorageMusicList(own_storage.OwnStorageMusicListRequest{
			SearchKeyword: req.SearchKeyword,
			Limit:         limit,
			Offset:        offset,
		}, db)

		if err != nil {
			finalResponse.Error = err
			ch <- finalResponse
			return
		}

		finalResponse.Response.TotalCount = resp.TotalCount
		if len(resp.Items) > 0 {
			var songs []internal.SongModel

			for _, song := range resp.Items {
				ext := "mp3"
				spl := strings.Split(song.Url, ".")
				if len(spl) >= 2 {
					ext = spl[len(spl)-1]
				}

				songs = append(songs, internal.SongModel{
					ExternalId: fmt.Sprint(song.Id),
					Title:      song.Title,
					Artist:     song.Artist,
					ImageUrl:   song.ImageUrl,
					Genre:      song.Genre,
					Duration:   song.Duration,
					Files: map[string]string{
						ext: song.Url,
					},
				})
			}

			finalResponse.Response.Songs = songs
		}

		ch <- finalResponse
	}()

	return ch
}
