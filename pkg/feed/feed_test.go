package feed

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/wrappers"
	"github.com/digitalmonsters/go-common/wrappers/follow"
	"github.com/digitalmonsters/go-common/wrappers/go_tokenomics"
	"github.com/digitalmonsters/go-common/wrappers/like"
	"github.com/digitalmonsters/go-common/wrappers/music"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/music/configs"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/digitalmonsters/music/pkg/feed/deduplicator"
	"github.com/digitalmonsters/music/pkg/feed/feed_converter"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.elastic.co/apm"
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

func TestNewFeed(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresTables(config.MasterDb, []string{"public.creator_songs", "public.categories", "public.moods"}, nil, t); err != nil {
		t.Fatal(err)
	}

	userId := int64(1)

	cat := addCategory(t, "test_cat")
	mood := addMood(t, "test_mood")

	var songs []database.CreatorSong

	for i := 1; i <= 10; i++ {
		songs = append(songs, database.CreatorSong{
			UserId:       userId,
			Name:         fmt.Sprintf("test song %v", i),
			Status:       music.CreatorSongStatusApproved,
			CreatedAt:    time.Now(),
			CategoryId:   cat.Id,
			MoodId:       mood.Id,
			ShortListens: i,
		})
	}

	if err := gormDb.Create(&songs).Error; err != nil {
		t.Fatal(err)
	}

	listenedMusic := []database.ListenedMusic{
		{
			UserId: userId,
			SongId: songs[0].Id,
		},
		{
			UserId: userId,
			SongId: songs[1].Id,
		},
	}

	if err := gormDb.Create(&listenedMusic).Error; err != nil {
		t.Fatal(err)
	}

	userWrapper := &user_go.UserGoWrapperMock{}

	userWrapper.GetUsersFn = func(userIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]user_go.UserRecord] {
		ch := make(chan wrappers.GenericResponseChan[map[int64]user_go.UserRecord], 2)
		resp := map[int64]user_go.UserRecord{}

		for _, userId := range userIds {
			resp[userId] = user_go.UserRecord{
				UserId:    userId,
				Username:  fmt.Sprintf("username%v", userId),
				Firstname: fmt.Sprintf("firstname%v", userId),
				Lastname:  fmt.Sprintf("lastname%v", userId),
				Email:     fmt.Sprintf("email%v", userId),
			}
		}

		ch <- wrappers.GenericResponseChan[map[int64]user_go.UserRecord]{
			Error:    nil,
			Response: resp,
		}
		close(ch)

		return ch
	}

	followWrapper := &follow.FollowWrapperMock{}

	followWrapper.GetUserFollowingRelationBulkFn = func(userId int64, requestUserIds []int64, apmTransaction *apm.Transaction,
		forceLog bool) chan follow.GetUserFollowingRelationBulkResponseChan {
		ch := make(chan follow.GetUserFollowingRelationBulkResponseChan, 2)

		ch <- follow.GetUserFollowingRelationBulkResponseChan{
			Error: nil,
			Data:  map[int64]follow.RelationData{},
		}
		close(ch)

		return ch
	}

	likeWrapper := &like.LikeWrapperMock{}

	likeWrapper.GetInternalSpotReactionsByUserFn = func(contentIds []int64, userId int64, apmTransaction *apm.Transaction, forceLog bool) chan like.GetInternalSpotReactionsByUserResponseChan {
		ch := make(chan like.GetInternalSpotReactionsByUserResponseChan, 2)

		ch <- like.GetInternalSpotReactionsByUserResponseChan{
			Error: nil,
			Data:  map[int64]like.SpotReaction{},
		}
		close(ch)

		return ch
	}

	goTokenomicsWrapper := &go_tokenomics.GoTokenomicsWrapperMock{}

	goTokenomicsWrapper.GetContentEarningsTotalByContentIdsFn = func(contentIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan go_tokenomics.GetContentEarningsTotalByContentIdsResponseChan {
		ch := make(chan go_tokenomics.GetContentEarningsTotalByContentIdsResponseChan, 2)

		ch <- go_tokenomics.GetContentEarningsTotalByContentIdsResponseChan{
			Error: nil,
			Items: map[int64]decimal.Decimal{},
		}
		close(ch)

		return ch
	}

	feedConverter := feed_converter.NewFeedConverter(userWrapper, followWrapper, likeWrapper, goTokenomicsWrapper, context.Background())
	deDuplicator := deduplicator.GetMock()
	var configurator = &application.Configurator[configs.AppConfig]{}
	configurator.Values.MUSIC_MAX_HASHTAGS_COUNT = 1000
	configurator.Values.MUSIC_FEED_LIMIT = 20000
	configurator.Values.MUSIC_CALCULATION_LOVE_COUNT_WEIGHT = 10
	configurator.Values.MUSIC_CALCULATION_LIKE_COUNT_WEIGHT = 6
	configurator.Values.MUSIC_CALCULATION_SHORT_LISTEN_COUNT_WEIGHT = 1
	configurator.Values.MUSIC_CALCULATION_DISLIKE_COUNT_WEIGHT = 1
	configurator.Values.MUSIC_CALCULATION_TIMING_START_CONF = 5000
	configurator.Values.MUSIC_CALCULATION_TIMING_DELIMITER = 60
	configurator.Values.MUSIC_FEED_UPDATE_SCORE_FREQUENCY_MINUTES = 60

	musicFeed := NewFeed(deDuplicator, feedConverter, nil, configurator)

	feedSongs, err := musicFeed.GetFeed(gormDb, userId, []int64{}, 10, router.MethodExecutionData{
		Context: context.Background(),
		UserId:  userId,
	})
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, feedSongs.Data, 8)
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

func TestFeedBuilder_LaunchTask(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresTables(config.MasterDb, []string{"public.creator_songs", "public.categories", "public.moods"}, nil, t); err != nil {
		t.Fatal(err)
	}

	var configurator = &application.Configurator[configs.AppConfig]{}
	configurator.Values.MUSIC_MAX_HASHTAGS_COUNT = 1000
	configurator.Values.MUSIC_FEED_LIMIT = 20000
	configurator.Values.MUSIC_CALCULATION_LOVE_COUNT_WEIGHT = 10
	configurator.Values.MUSIC_CALCULATION_LIKE_COUNT_WEIGHT = 6
	configurator.Values.MUSIC_CALCULATION_SHORT_LISTEN_COUNT_WEIGHT = 1
	configurator.Values.MUSIC_CALCULATION_DISLIKE_COUNT_WEIGHT = 1
	configurator.Values.MUSIC_CALCULATION_TIMING_START_CONF = 5000
	configurator.Values.MUSIC_CALCULATION_TIMING_DELIMITER = 60
	configurator.Values.MUSIC_FEED_UPDATE_SCORE_FREQUENCY_MINUTES = 60

	userId := int64(1)
	cat := addCategory(t, "test_cat")
	mood := addMood(t, "test_mood")

	var songs []database.CreatorSong

	for i := 1; i <= 100; i++ {
		songs = append(songs, database.CreatorSong{
			UserId:       userId,
			Name:         fmt.Sprintf("test song %v", i),
			Status:       music.CreatorSongStatusApproved,
			CreatedAt:    time.Now(),
			CategoryId:   cat.Id,
			MoodId:       mood.Id,
			ShortListens: i,
			Likes:        i,
			Loves:        i,
			Dislikes:     i,
		})
	}

	if err := gormDb.Create(&songs).Error; err != nil {
		t.Fatal(err)
	}

	builder := newMusicFeedBuilder(nil, gormDb, configurator)

	if err := builder.updateMusicFeed(gormDb, context.Background()); err != nil {
		t.Fatal(err)
	}

	var records []database.CreatorSong
	if err := gormDb.Order("score desc").Find(&records).Error; err != nil {
		t.Fatal(err)
	}

	assert.Len(t, records, 100)
	assert.Equal(t, records[0].Id, songs[len(records)-1].Id)
	assert.Equal(t, records[len(records)-1].Id, songs[0].Id)
}

func TestNewFeed2(t *testing.T) {
	var finalRespItems []MusicFeedItem

	b, _ := json.Marshal(finalRespItems)
	fmt.Println(string(b))
}
