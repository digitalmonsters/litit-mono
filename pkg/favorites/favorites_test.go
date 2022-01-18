package favorites

import (
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/digitalmonsters/music/configs"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"os"
	"testing"
	"time"
)

var config configs.Settings
var gormDb *gorm.DB

func TestMain(m *testing.M) {
	config = configs.GetConfig()
	gormDb = database.GetDb(database.DbTypeMaster)
	os.Exit(m.Run())
}

func TestAddToFavorites(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresTables(config.MasterDb, []string{"public.favorites", "public.songs"}, nil, t); err != nil {
		t.Fatal(err)
	}
	for i := int64(1); i <= 5; i++ {
		testName := fmt.Sprintf("test%v", i)
		song := database.Song{
			Source:     database.SongSourceSoundStripe,
			ExternalId: testName,
			Title:      testName,
			Artist:     testName,
			ImageUrl:   testName,
		}
		if err := gormDb.Create(&song).Error; err != nil {
			t.Fatal(err)
		}
		if err := AddToFavorites(AddToFavoritesRequest{
			UserId: i,
			SongId: song.Id,
		}, gormDb); err != nil {
			t.Fatal(err)
		}
	}

	var favorites []database.Favorite

	if err := gormDb.Find(&favorites).Error; err != nil {
		t.Fatal(err)
	}

	a := assert.New(t)

	a.Equal(5, len(favorites))

	for _, favorite := range favorites {
		a.Equal(favorite.SongId, favorite.SongId)
	}
}

func TestRemoveFromFavorites(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresTables(config.MasterDb, []string{"public.favorites", "public.songs"}, nil, t); err != nil {
		t.Fatal(err)
	}

	var songs []database.Song
	for i := int64(1); i <= 5; i++ {
		testName := fmt.Sprintf("test%v", i)
		song := database.Song{
			Source:     database.SongSourceSoundStripe,
			ExternalId: testName,
			Title:      testName,
			Artist:     testName,
			ImageUrl:   testName,
		}
		if err := gormDb.Create(&song).Error; err != nil {
			t.Fatal(err)
		}

		songs = append(songs, song)

		favorite := database.Favorite{
			UserId: song.Id,
			SongId: song.Id,
		}
		if err := gormDb.Create(&favorite).Error; err != nil {
			t.Fatal(err)
		}
	}

	if err := RemoveFromFavorites(RemoveFromFavoritesRequest{
		UserId: songs[0].Id,
		SongId: songs[0].Id,
	}, gormDb); err != nil {
		t.Fatal(err)
	}

	var favorites []database.Favorite

	if err := gormDb.Find(&favorites).Error; err != nil {
		t.Fatal(err)
	}

	a := assert.New(t)

	a.Equal(4, len(favorites))
}

func TestFavoriteSongsList(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresTables(config.MasterDb, []string{"public.favorites", "public.songs"}, nil, t); err != nil {
		t.Fatal(err)
	}

	userId := 222

	for i := int64(1); i <= 5; i++ {
		testName := fmt.Sprintf("test%v", i)
		song := database.Song{
			Source:     database.SongSourceSoundStripe,
			ExternalId: testName,
			Title:      testName,
			Artist:     testName,
			ImageUrl:   testName,
		}
		if err := gormDb.Create(&song).Error; err != nil {
			t.Fatal(err)
		}
		favorite := database.Favorite{
			UserId:    int64(userId),
			SongId:    song.Id,
			CreatedAt: time.Date(2021, 5, int(i), 1, 1, 1, 1, time.UTC),
		}
		if err := gormDb.Create(&favorite).Error; err != nil {
			t.Fatal(err)
		}
	}

	resp, err := FavoriteSongsList(FavoriteSongsListRequest{
		Count:  2,
		Cursor: "",
	}, int64(userId), gormDb)
	assert.Nil(t, err)
	assert.Len(t, resp.Songs, 2)
}
