package own_storage

import (
	"fmt"
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

func TestUpsertSongsToOwnStorageBulk(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresTables(config.MasterDb, []string{"public.music_storage"}, nil, t); err != nil {
		t.Fatal(err)
	}

	var items []OwnSongItem

	for i := 0; i < 5; i++ {
		str := fmt.Sprintf("test_%v", i)
		items = append(items, OwnSongItem{
			Title:       str,
			Description: str,
			Artist:      str,
			ImageUrl:    str,
			FileUrl:     str,
			Genre:       str,
			Duration:    20,
		})
	}

	resp, err := UpsertSongsToOwnStorageBulk(AddSongsToOwnStorageRequest{Items: items}, gormDb)
	assert.Nil(t, err)
	assert.Len(t, resp, len(items))

	hardcode := "new_value"
	resp, err = UpsertSongsToOwnStorageBulk(AddSongsToOwnStorageRequest{Items: []OwnSongItem{
		{
			Id:    null.IntFrom(resp[0].Id),
			Title: hardcode,
		},
	}}, gormDb)

	assert.Nil(t, err)
	assert.Len(t, resp, 1)

	var songs []database.MusicStorage
	err = gormDb.Model(songs).Order("id asc").Find(&songs).Error
	assert.Nil(t, err)
	assert.Len(t, songs, 5)
	assert.Equal(t, songs[0].Title, hardcode)
}

func TestDeleteSongsFromOwnStorageBulk(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresTables(config.MasterDb, []string{"public.music_storage"}, nil, t); err != nil {
		t.Fatal(err)
	}

	var items []OwnSongItem

	for i := 0; i < 5; i++ {
		str := fmt.Sprintf("test_%v", i)
		items = append(items, OwnSongItem{
			Title:       str,
			Description: str,
			Artist:      str,
			ImageUrl:    str,
			FileUrl:     str,
			Genre:       str,
			Duration:    20,
		})
	}

	resp, err := UpsertSongsToOwnStorageBulk(AddSongsToOwnStorageRequest{Items: items}, gormDb)
	assert.Nil(t, err)
	assert.Len(t, resp, len(items))

	err = DeleteSongsFromOwnStorageBulk(DeleteSongsFromOwnStorageRequest{SongIds: []int64{resp[0].Id}}, gormDb)
	assert.Nil(t, err)

	var songs []database.MusicStorage
	err = gormDb.Model(songs).Order("id asc").Find(&songs).Error
	assert.Nil(t, err)
	assert.Len(t, songs, 4)
}

func TestOwnStorageMusicList(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresTables(config.MasterDb, []string{"public.music_storage"}, nil, t); err != nil {
		t.Fatal(err)
	}

	var items []OwnSongItem

	for i := 0; i < 5; i++ {
		str := fmt.Sprintf("test_%v", i)
		items = append(items, OwnSongItem{
			Title:       str,
			Description: str,
			Artist:      str,
			ImageUrl:    str,
			FileUrl:     str,
			Genre:       str,
			Duration:    20,
		})
	}

	resp, err := UpsertSongsToOwnStorageBulk(AddSongsToOwnStorageRequest{Items: items}, gormDb)
	assert.Nil(t, err)
	assert.Len(t, resp, len(items))

	list, err := OwnStorageMusicList(OwnStorageMusicListRequest{
		SearchKeyword: null.StringFrom("3"),
		Order:         0,
		Limit:         10,
		Offset:        0,
	}, gormDb)
	assert.Nil(t, err)

	assert.Len(t, list.Items, 1)
	fmt.Println(list.Items)

}
