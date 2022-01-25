package frontend

import (
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/pkg/errors"
	"github.com/thoas/go-funk"
	"go.elastic.co/apm"
	"gorm.io/gorm"
)

func ConvertSongsToFrontendModel(songs []database.Song, userId int64, db *gorm.DB, apmTransaction *apm.Transaction) []Song {
	songsArr := make([]*Song, 0)

	for _, song := range songs {
		songsArr = append(songsArr, &Song{
			Id:       song.Id,
			Title:    song.Title,
			Artist:   song.Artist,
			Url:      song.Artist,
			ImageUrl: song.ImageUrl,
			Genre:    song.Genre,
			Duration: song.Duration,
		})
	}

	songsMapped := map[int64]*Song{}
	songsIds := make([]int64, 0)
	for _, s := range songsArr {
		songsMapped[s.Id] = s
		songsIds = append(songsIds, s.Id)
	}

	routines := []chan error{
		fillIsFavorite(songsMapped, userId, songsIds, db),
	}

	for _, c := range routines {
		if err := <-c; err != nil {
			apm_helper.CaptureApmError(err, apmTransaction)
		}
	}

	var result []Song
	for _, s := range songsArr {
		result = append(result, *s)
	}

	return result
}

func ConvertPlaylistsToFrontendModel(playlists []database.Playlist) (result []Playlist) {
	for _, pl := range playlists {
		result = append(result, Playlist{
			Id:         pl.Id,
			Name:       pl.Name,
			Color:      pl.Color,
			SongsCount: pl.SongsCount,
		})
	}

	return result
}

func fillIsFavorite(songsMapped map[int64]*Song, currentUserId int64, songIds []int64, db *gorm.DB) chan error {
	ch := make(chan error, 2)

	go func() {
		defer func() {
			close(ch)
		}()

		if currentUserId == 0 {
			return
		}

		var favoriteSongIds []int64

		if err := db.Table("favorites").Where("favorites.user_id = ? and favorites.song_id in ?", currentUserId, songIds).
			Select("favorites.song_id").Find(&favoriteSongIds).Error; err != nil {
			ch <- errors.Wrap(err, "fill is following")
		}

		if len(favoriteSongIds) > 0 {
			for _, m := range songsMapped {
				if funk.ContainsInt64(favoriteSongIds, m.Id) {
					m.IsFavorite = true
				}
			}
		}
	}()

	return ch
}
