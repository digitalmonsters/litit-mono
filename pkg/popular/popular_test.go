package popular

import (
	"fmt"
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
	config = configs.GetConfig()
	gormDb = database.GetDb(database.DbTypeMaster)
	os.Exit(m.Run())
}

func TestGetPopularSongs(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresTables(config.MasterDb, []string{"public.playlists", "public.songs", "public.playlist_song_relations"}, nil, t); err != nil {
		t.Fatal(err)
	}

	var songs []database.Song
	for i := 0; i < 50; i++ {
		songs = append(songs, database.Song{
			ExternalId:   fmt.Sprintf("test_%v", i),
			Source:       database.SongSourceSoundStripe,
			Title:        fmt.Sprintf("test_%v", i),
			Artist:       fmt.Sprintf("artist_%v", i),
			ListenAmount: i * 100,
		})
	}

	err := gormDb.Create(&songs).Error
	assert.Nil(t, err)

	resp, err := GetPopularSongs(GetPopularSongsRequest{
		Count: 10,
	}, gormDb)
	assert.Nil(t, err)
	assert.Len(t, resp.Songs, 10)
}
