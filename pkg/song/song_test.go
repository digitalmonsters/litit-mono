package song

import (
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/digitalmonsters/music/configs"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/digitalmonsters/music/pkg/music_source"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"os"
	"testing"
)

var config configs.Settings
var gormDb *gorm.DB
var musicStorageService *music_source.MusicStorageService

func TestMain(m *testing.M) {
	config = configs.GetConfig()
	gormDb = database.GetDb(database.DbTypeMaster)
	musicStorageService = music_source.NewMusicStorageService(&config)

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
		Source:     database.SongSourceSoundStripe,
		ExternalId: "123",
		Title:      "test",
		Artist:     "artist",
	}

	err = gormDb.Create(&song).Error
	assert.Nil(t, err)

	//err = AddSongToPlaylistBulk(AddSongToPlaylistRequest{
	//	Songs: []RelationItem{
	//		{
	//			SongId:     song.Id,
	//			PlaylistId: playlist.Id,
	//			SortOrder:  1,
	//		},
	//	},
	//}, gormDb, nil, soundStripeService)
	//assert.Nil(t, err)

	if err := gormDb.Create(&database.PlaylistSongRelations{
		PlaylistId: playlist.Id,
		SongId:     song.Id,
		SortOrder:  1,
	}).Error; err != nil {
		t.Fatal(err)
	}

	var relation database.PlaylistSongRelations
	err = gormDb.Where("song_id = ? and playlist_id = ?", song.Id, playlist.Id).First(&relation).Error
	assert.Nil(t, err)

	err = gormDb.First(&playlist, playlist.Id).Error
	assert.Nil(t, err)

	assert.Equal(t, playlist.SongsCount, 1)

	err = DeleteSongFromPlaylistsBulk(DeleteSongsFromPlaylistBulkRequest{
		PlaylistId: playlist.Id,
		SongsIds:   []int64{song.Id},
	}, gormDb)
	assert.Nil(t, err)

	err = gormDb.Where("song_id = ? and playlist_id = ?", song.Id, playlist.Id).First(&relation).Error
	assert.NotNil(t, err)

	err = gormDb.First(&playlist, playlist.Id).Error
	assert.Nil(t, err)

	assert.Equal(t, playlist.SongsCount, 0)
}

func TestPlaylistSongList(t *testing.T) {
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
		Source:     database.SongSourceSoundStripe,
		ExternalId: "123",
		Title:      "test",
		Artist:     "artist",
	}

	err = gormDb.Create(&song).Error
	assert.Nil(t, err)

	song2 := database.Song{
		Source:     database.SongSourceSoundStripe,
		ExternalId: "234",
		Title:      "test2",
		Artist:     "artist2",
	}

	err = gormDb.Create(&song2).Error
	assert.Nil(t, err)

	songRelations := []database.PlaylistSongRelations{
		{
			PlaylistId: playlist.Id,
			SongId:     song.Id,
			SortOrder:  1,
		},
		{
			PlaylistId: playlist.Id,
			SongId:     song2.Id,
			SortOrder:  2,
		},
	}

	err = gormDb.Create(&songRelations).Error
	assert.Nil(t, err)

	_, err = PlaylistSongListAdmin(PlaylistSongListRequest{
		PlaylistId: playlist.Id,
		Limit:      1,
		Offset:     0,
	}, gormDb)
	assert.Nil(t, err)
}
