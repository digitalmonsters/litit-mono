package playlist

import (
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
	var err error
	config = configs.GetConfig()
	gormDb, err = boilerplate_testing.GetPostgresConnection(&config.MasterDb)
	if err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func TestUpsertPlaylist(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresTables(config.MasterDb, []string{"public.playlists"}, nil, t); err != nil {
		t.Fatal(err)
	}

	playlistReq := UpsertPlaylistRequest{
		Name:      "Test",
		SortOrder: 1,
		Color:     "#ff0000",
	}

	resp, err := UpsertPlaylist(playlistReq, gormDb)
	assert.Nil(t, err)

	var playlist database.Playlist
	err = gormDb.First(&playlist, resp.Id).Error
	assert.Nil(t, err)
	assert.Equal(t, playlist.Name, playlistReq.Name)
	assert.Equal(t, playlist.SortOrder, playlistReq.SortOrder)
	assert.Equal(t, playlist.Color, playlistReq.Color)

	playlistReq = UpsertPlaylistRequest{
		Id:        null.IntFrom(playlist.Id),
		Name:      "Test2",
		SortOrder: 2,
		Color:     "#ff1111",
	}

	resp, err = UpsertPlaylist(playlistReq, gormDb)
	assert.Nil(t, err)

	err = gormDb.First(&playlist, resp.Id).Error
	assert.Nil(t, err)
	assert.Equal(t, playlist.Name, playlistReq.Name)
	assert.Equal(t, playlist.SortOrder, playlistReq.SortOrder)
	assert.Equal(t, playlist.Color, playlistReq.Color)
}

func TestDeletePlaylistsBulk(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresTables(config.MasterDb, []string{"public.playlists"}, nil, t); err != nil {
		t.Fatal(err)
	}

	playlistsToAdd := []database.Playlist{
		{
			Name:      "test1",
			SortOrder: 1,
			Color:     "test1",
		},
		{
			Name:      "test2",
			SortOrder: 2,
			Color:     "test2",
		},
		{
			Name:      "test3",
			SortOrder: 3,
			Color:     "test3",
		},
	}

	err := gormDb.Create(&playlistsToAdd).Error
	assert.Nil(t, err)

	err = DeletePlaylistsBulk(DeletePlaylistsBulkRequest{Ids: []int64{playlistsToAdd[0].Id, playlistsToAdd[1].Id}}, gormDb)
	assert.Nil(t, err)

	var playlists []database.Playlist
	err = gormDb.Find(&playlists).Error
	assert.Nil(t, err)
	assert.Equal(t, len(playlists), 1)
	assert.Equal(t, playlists[0].Id, playlistsToAdd[2].Id)
}
