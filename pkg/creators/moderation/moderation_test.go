package moderation

import (
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/music/configs"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/stretchr/testify/assert"
	"go.elastic.co/apm"
	"gorm.io/gorm"
	"os"
	"testing"
)

var config configs.Settings
var gormDb *gorm.DB
var userGoWrapper *user_go.UserGoWrapperMock

func TestMain(m *testing.M) {
	config = configs.GetConfig()
	gormDb = database.GetDb(database.DbTypeMaster)
	os.Exit(m.Run())
}

func addCategory(t *testing.T, categoryName string) *database.Category {
	catergory := database.Category{
		Name: categoryName,
	}

	if err := gormDb.Create(&catergory).Error; err != nil {
		t.Fatal(err)
	}

	userGoWrapper = &user_go.UserGoWrapperMock{}

	userGoWrapper.GetUsersFn = func(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan user_go.GetUsersResponseChan {
		ch := make(chan user_go.GetUsersResponseChan, 2)
		defer close(ch)

		mapped := map[int64]user_go.UserRecord{}
		for _, id := range userIds {
			mapped[id] = user_go.UserRecord{
				UserId:   id,
				Username: fmt.Sprintf("user %v", id),
			}
		}

		ch <- user_go.GetUsersResponseChan{
			Error: nil,
			Items: mapped,
		}

		return ch
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

func TestRejectMusic(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresTables(config.MasterDb, []string{
		"public.creator_reject_reasons",
		"public.creator_songs",
		"public.categories",
		"public.moods",
	}, nil, t); err != nil {
		t.Fatal(err)
	}

	mood := addMood(t, "test_mood")
	category := addCategory(t, "test_category")

	rejectReasons := []database.CreatorRejectReasons{
		{
			Type:   database.ReasonTypeCreatorSong,
			Reason: "test1",
		},
		{
			Type:   database.ReasonTypeCreatorSong,
			Reason: "test2",
		},
	}

	if err := gormDb.Create(&rejectReasons).Error; err != nil {
		t.Fatal(err)
	}

	song := database.CreatorSong{
		UserId:     1,
		Name:       "test_song",
		MoodId:     mood.Id,
		CategoryId: category.Id,
		Status:     database.CreatorSongStatusPublished,
	}

	if err := gormDb.Create(&song).Error; err != nil {
		t.Fatal(err)
	}

	err := RejectMusic(RejectMusicRequest{
		SongId:       song.Id,
		RejectReason: rejectReasons[0].Id,
	}, gormDb)

	assert.Nil(t, err)

	var songToCheck database.CreatorSong
	if err = gormDb.Preload("Reject").First(&songToCheck, song.Id).Error; err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, songToCheck.Status, database.CreatorSongStatusRejected)
	assert.Equal(t, songToCheck.RejectReason.Int64, rejectReasons[0].Id)
	assert.Equal(t, songToCheck.Reject.Reason, "test1")

	err = ApproveMusic(ApproveMusicRequest{
		SongId: songToCheck.Id,
	}, gormDb)

	assert.Nil(t, err)

	if err = gormDb.Preload("Reject").First(&songToCheck, song.Id).Error; err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, songToCheck.Status, database.CreatorSongStatusApproved)
	assert.False(t, songToCheck.RejectReason.Valid)
}

func TestList(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresTables(config.MasterDb, []string{
		"public.creator_reject_reasons",
		"public.creator_songs",
		"public.categories",
		"public.moods",
	}, nil, t); err != nil {
		t.Fatal(err)
	}

	mood := addMood(t, "test_mood")
	category := addCategory(t, "test_category")

	var songs []database.CreatorSong
	for i := 0; i < 10; i++ {
		songs = append(songs, database.CreatorSong{
			UserId:     1,
			Name:       fmt.Sprintf("test %v", i),
			Status:     database.CreatorSongStatusPublished,
			CategoryId: category.Id,
			MoodId:     mood.Id,
		})
	}

	if err := gormDb.Create(&songs).Error; err != nil {
		t.Fatal(err)
	}

	resp, err := List(ListRequest{Limit: 20}, gormDb, userGoWrapper, nil)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(resp.Items), 10)
}
