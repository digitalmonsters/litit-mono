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
	var err error
	config = configs.GetConfig()
	gormDb, err = boilerplate_testing.GetPostgresConnection(&config.MasterDb)
	if err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func TestAddToFavorites(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresTables(config.MasterDb, []string{"public.favorites", "public.songs"}, nil, t); err != nil {
		t.Fatal(err)
	}
	for i := int64(1); i <= 5; i++ {
		testName := fmt.Sprintf("test%v", i)
		song := database.Song{
			Id:       testName,
			Title:    testName,
			Artist:   testName,
			Url:      testName,
			ImageUrl: testName,
		}
		if err := gormDb.Create(&song).Error; err != nil {
			t.Fatal(err)
		}
		if err := AddToFavorites(AddToFavoritesRequest{
			UserId: i,
			SongId: testName,
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
		a.Equal(fmt.Sprintf("test%v", favorite.UserId), favorite.SongId)
	}
}

func TestRemoveFromFavorites(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresTables(config.MasterDb, []string{"public.favorites", "public.songs"}, nil, t); err != nil {
		t.Fatal(err)
	}
	for i := int64(1); i <= 5; i++ {
		testName := fmt.Sprintf("test%v", i)
		song := database.Song{
			Id:       testName,
			Title:    testName,
			Artist:   testName,
			Url:      testName,
			ImageUrl: testName,
		}
		if err := gormDb.Create(&song).Error; err != nil {
			t.Fatal(err)
		}
		favorite := database.Favorite{
			UserId: i,
			SongId: song.Id,
		}
		if err := gormDb.Create(&favorite).Error; err != nil {
			t.Fatal(err)
		}
	}

	if err := RemoveFromFavorites(RemoveFromFavoritesRequest{
		UserId: 1,
		SongId: "test1",
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
			Id:       testName,
			Title:    testName,
			Artist:   testName,
			Url:      testName,
			ImageUrl: testName,
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
