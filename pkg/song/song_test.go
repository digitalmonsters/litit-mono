package song

import (
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/digitalmonsters/music/configs"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/stretchr/testify/assert"
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

func TestMethods(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresTables(config.MasterDb, []string{"public.playlists", "public.songs", "public.playlist_song_relations"}, nil, t); err != nil {
		t.Fatal(err)
	}

	playlist := database.Playlist{
		Name:      "test",
		SortOrder: 1,
	}

	err := gormDb.Create(&playlist).Error
	assert.Nil(t, err)

	song := database.Song{
		Id:     "test_song",
		Title:  "test",
		Artist: "artist",
	}

	err = gormDb.Create(&song).Error
	assert.Nil(t, err)

	err = AddSongToPlaylistBulk(AddSongToPlaylistRequest{
		Songs: []RelationItem{
			{
				SongId:     song.Id,
				PlaylistId: playlist.Id,
				SortOrder:  1,
			},
		},
	}, gormDb)
	assert.Nil(t, err)

	var relation database.PlaylistSongRelations
	err = gormDb.Where("song_id = ? and playlist_id = ?", song.Id, playlist.Id).First(&relation).Error
	assert.Nil(t, err)

	err = gormDb.First(&playlist, playlist.Id).Error
	assert.Nil(t, err)

	assert.Equal(t, playlist.SongsCount, 1)

	err = DeleteSongFromPlaylistsBulk(DeleteSongsFromPlaylistBulkRequest{
		PlaylistId: playlist.Id,
		SongsIds:   []string{song.Id},
	}, gormDb)
	assert.Nil(t, err)

	err = gormDb.Where("song_id = ? and playlist_id = ?", song.Id, playlist.Id).First(&relation).Error
	assert.NotNil(t, err)

	err = gormDb.First(&playlist, playlist.Id).Error
	assert.Nil(t, err)

	assert.Equal(t, playlist.SongsCount, 0)
}
