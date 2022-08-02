package listen

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/digitalmonsters/go-common/wrappers/music"
	"github.com/digitalmonsters/music/configs"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"os"
	"testing"
)

var counterService *listenCounterService
var config configs.Settings
var gormDb *gorm.DB

func TestMain(m *testing.M) {
	config = configs.GetConfig()
	gormDb = database.GetDb(database.DbTypeMaster)
	counterService = newListenCounterService(10)

	os.Exit(m.Run())
}

func TestProcessUpdateCounters(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresTables(config.MasterDb, []string{
		"public.creator_songs",
		"public.categories",
		"public.moods",
	}, nil, t); err != nil {
		t.Fatal(err)
	}

	mood := addMood(t, "test_mood")
	category := addCategory(t, "test_category")

	song := database.CreatorSong{
		UserId:     1,
		Name:       fmt.Sprintf("test %v", 1),
		Status:     music.CreatorSongStatusPublished,
		CategoryId: category.Id,
		MoodId:     mood.Id,
	}

	if err := gormDb.Create(&song).Error; err != nil {
		t.Fatal(err)
	}

	testMap := make(map[int64]*listenCount)
	testMap[song.Id] =
		&listenCount{
			legacyEvent: legacyEvent{
				Id:                song.Id,
				ListensCount:      song.UserId + 50,
				ShortListensCount: song.UserId + 100,
			},
			Messages: []kafka.Message{
				{
					Key: []byte(fmt.Sprint(song.Id)),
				},
			},
		}

	a := assert.New(t)
	var songs []database.CreatorSong
	if err := gormDb.Table("creator_songs").Find(&songs).Error; err != nil {
		t.Fatal(err)
	}

	for _, c := range songs {
		a.Equal(0, c.ShortListens)
		a.Equal(0, c.FullListens)
	}

	processed := counterService.Process(testMap, gormDb, nil, context.Background())
	a.Equal(len(processed), 1)

	if err := gormDb.Table("creator_songs").Find(&songs).Error; err != nil {
		t.Fatal(err)
	}
	for _, c := range songs {
		a.Equal(int(c.UserId+50), c.FullListens)
		a.Equal(int(c.UserId+100), c.ShortListens)
	}
}

func addCategory(t *testing.T, categoryName string) *database.Category {
	catergory := database.Category{
		Name: categoryName,
	}

	if err := gormDb.Create(&catergory).Error; err != nil {
		t.Fatal(err)
	}

	return &catergory
}

func addMood(t *testing.T, name string) *database.Mood {
	mood := database.Mood{
		Name: name,
	}

	if err := gormDb.Create(&mood).Error; err != nil {
		t.Fatal(err)
	}

	return &mood
}
