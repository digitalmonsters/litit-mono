package music_source

import (
	"context"
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/digitalmonsters/music/configs"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"os"
	"testing"
)

var config configs.Settings
var gormDb *gorm.DB

func TestMain(m *testing.M) {
	config = configs.GetConfig()
	gormDb = database.GetDb(database.DbTypeMaster)
	os.Exit(m.Run())
}

func TestNewMusicStorageService(t *testing.T) {
	service := NewMusicStorageService(&config)

	if err := boilerplate_testing.FlushPostgresTables(config.MasterDb, []string{"public.playlists", "public.songs", "public.playlist_song_relations"}, nil, t); err != nil {
		t.Fatal(err)
	}

	playlists := []database.Playlist{
		{
			Name:      "test",
			SortOrder: 1,
		},
		{
			Name:      "test2",
			SortOrder: 1,
		},
	}

	err := gormDb.Create(&playlists).Error
	assert.Nil(t, err)

	song := database.Song{
		ExternalId: "test222",
		Source:     database.SongSourceOwnStorage,
		Title:      "test",
		Artist:     "artist",
	}

	err = gormDb.Create(&song).Error
	assert.Nil(t, err)

	song2 := database.Song{
		ExternalId: "test123",
		Source:     database.SongSourceOwnStorage,
		Title:      "test2",
		Artist:     "artist2",
	}

	err = gormDb.Create(&song2).Error
	assert.Nil(t, err)

	songRelations := []database.PlaylistSongRelations{
		{
			PlaylistId: playlists[0].Id,
			SongId:     song.Id,
			SortOrder:  1,
		},
		{
			PlaylistId: playlists[1].Id,
			SongId:     song2.Id,
			SortOrder:  2,
		},
	}

	err = gormDb.Create(&songRelations).Error
	assert.Nil(t, err)

	resp, err := service.ListMusic(ListMusicRequest{
		SearchKeyword: null.String{},
		PlaylistIds:   []int64{songRelations[0].PlaylistId},
		Source:        database.SongSourceOwnStorage,
		Page:          1,
		Size:          10,
	}, gormDb, nil, context.TODO())
	assert.Nil(t, err)

	assert.Len(t, resp.Songs, 1)
	assert.Equal(t, resp.Songs[0].ExternalId, song.ExternalId)
}
