package creators

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/music/configs"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/stretchr/testify/assert"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"os"
	"testing"
	"time"
)

var config configs.Settings
var gormDb *gorm.DB
var userWrapper *user_go.UserGoWrapperMock
var service *Service

func TestMain(m *testing.M) {
	config = configs.GetConfig()
	gormDb = database.GetDb(database.DbTypeMaster)
	userWrapper = &user_go.UserGoWrapperMock{}
	service = NewService(nil)

	userWrapper.GetUsersFn = func(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan user_go.GetUsersResponseChan {
		ch := make(chan user_go.GetUsersResponseChan, 2)
		resp := map[int64]user_go.UserRecord{}

		for _, userId := range userIds {
			resp[userId] = user_go.UserRecord{
				UserId: userId,
			}
		}

		ch <- user_go.GetUsersResponseChan{
			Error: nil,
			Items: resp,
		}
		close(ch)

		return ch
	}

	os.Exit(m.Run())
}

func TestBecomeMusicCreator(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresTables(config.MasterDb, []string{"public.creators"}, nil, t); err != nil {
		t.Fatal(err)
	}

	userId := int64(1)
	link := "https://music"
	executionData := router.MethodExecutionData{
		Context: context.TODO(),
		UserId:  userId,
	}

	err := service.BecomeMusicCreator(BecomeMusicCreatorRequest{LibraryLink: link}, gormDb, executionData)
	assert.Nil(t, err)

	var creatorRequest database.Creator
	err = gormDb.Model(creatorRequest).Where("user_id = ?", userId).Find(&creatorRequest).Error
	assert.Nil(t, err)
	assert.Equal(t, creatorRequest.LibraryUrl, link)

	err = service.BecomeMusicCreator(BecomeMusicCreatorRequest{LibraryLink: link}, gormDb, executionData)
	assert.NotNil(t, err)
}

func TestCreatorRequestsList(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresTables(config.MasterDb, []string{"public.creators"}, nil, t); err != nil {
		t.Fatal(err)
	}

	var requests []database.Creator
	for i := 0; i < 10; i++ {
		r := database.Creator{
			Status:     eventsourcing.CreatorStatusPending,
			UserId:     int64(i),
			LibraryUrl: fmt.Sprintf("https://music%v", i),
		}

		if i%2 == 0 {
			r.Status = eventsourcing.CreatorStatusApproved
			r.ApprovedAt = null.TimeFrom(time.Now())
		}

		requests = append(requests, r)
	}

	if err := gormDb.Create(&requests).Error; err != nil {
		t.Fatal(err)
	}

	userId := int64(100)
	thresholdRequest := database.Creator{
		UserId:    userId,
		Status:    eventsourcing.CreatorStatusPending,
		CreatedAt: time.Date(2020, 1, 1, 1, 1, 1, 1, time.UTC),
	}

	if err := gormDb.Create(&thresholdRequest).Error; err != nil {
		t.Fatal(err)
	}

	resp, err := service.CreatorRequestsList(CreatorRequestsListRequest{
		Statuses:             []eventsourcing.CreatorStatus{eventsourcing.CreatorStatusPending},
		MaxThresholdExceeded: false,
		Limit:                10,
		Offset:               0,
	}, gormDb, config.Creators.MaxThresholdHours, nil, userWrapper)

	assert.Nil(t, err)
	assert.Len(t, resp.Items, 6)

	resp, err = service.CreatorRequestsList(CreatorRequestsListRequest{
		MaxThresholdExceeded: true,
		Limit:                10,
		Offset:               0,
	}, gormDb, config.Creators.MaxThresholdHours, nil, userWrapper)

	assert.Nil(t, err)
	assert.Len(t, resp.Items, 1)
	assert.Equal(t, resp.Items[0].UserId, userId)
}

func TestCreatorRequestApprove(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresTables(config.MasterDb, []string{"public.creators"}, nil, t); err != nil {
		t.Fatal(err)
	}

	var requests []database.Creator
	for i := 0; i < 10; i++ {
		r := database.Creator{
			Status:     eventsourcing.CreatorStatusPending,
			UserId:     int64(i),
			LibraryUrl: fmt.Sprintf("https://music%v", i),
		}

		if i%2 == 0 {
			r.Status = eventsourcing.CreatorStatusApproved
			r.ApprovedAt = null.TimeFrom(time.Now())
		}

		requests = append(requests, r)
	}

	if err := gormDb.Create(&requests).Error; err != nil {
		t.Fatal(err)
	}

	var ids []int64
	for _, r := range requests {
		ids = append(ids, r.Id)
	}

	_, err := service.CreatorRequestApprove(CreatorRequestApproveRequest{Ids: ids}, gormDb)
	assert.Nil(t, err)

	var finalRequests []database.Creator
	err = gormDb.Find(&finalRequests).Error
	assert.Nil(t, err)

	for _, r := range finalRequests {
		assert.Equal(t, r.Status, eventsourcing.CreatorStatusApproved)
	}
}

func TestCreatorRequestReject(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresTables(config.MasterDb, []string{"public.creators", "public.creator_reject_reasons"}, nil, t); err != nil {
		t.Fatal(err)
	}

	var requests []database.Creator
	for i := 0; i < 10; i++ {
		r := database.Creator{
			Status:     eventsourcing.CreatorStatusPending,
			UserId:     int64(i),
			LibraryUrl: fmt.Sprintf("https://music%v", i),
		}

		if i%2 == 0 {
			r.Status = eventsourcing.CreatorStatusRejected
			r.ApprovedAt = null.TimeFrom(time.Now())
		}

		requests = append(requests, r)
	}

	if err := gormDb.Create(&requests).Error; err != nil {
		t.Fatal(err)
	}

	reasons := []database.CreatorRejectReasons{
		{
			Reason: "test1",
		},
		{
			Reason: "test2",
		},
	}

	if err := gormDb.Create(&reasons).Error; err != nil {
		t.Fatal(err)
	}

	var items []creatorRejectItem
	for _, r := range requests {
		items = append(items, creatorRejectItem{
			Id:     r.Id,
			Reason: reasons[0].Id,
		})
	}

	_, err := service.CreatorRequestReject(CreatorRequestRejectRequest{Items: items}, gormDb)
	assert.Nil(t, err)

	var finalRequests []database.Creator
	err = gormDb.Find(&finalRequests).Error
	assert.Nil(t, err)

	for _, r := range finalRequests {
		assert.Equal(t, r.Status, eventsourcing.CreatorStatusRejected)
	}
}

func addCreator(t *testing.T, userId int64, status eventsourcing.CreatorStatus) *database.Creator {
	creator := database.Creator{
		UserId:     userId,
		Status:     status,
		LibraryUrl: "https://test.com",
	}

	if err := gormDb.Create(&creator).Error; err != nil {
		t.Fatal(err)
	}

	return &creator
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

func TestUploadNewSong(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresTables(config.MasterDb, []string{"" +
		"public.creators", "public.categories", "public.creator_songs", "public.moods"}, nil, t); err != nil {
		t.Fatal(err)
	}

	userId := int64(111)
	creator := addCreator(t, userId, eventsourcing.CreatorStatusApproved)
	assert.NotNil(t, creator)

	executionData := router.MethodExecutionData{
		ApmTransaction: nil,
		Context:        context.TODO(),
		UserId:         userId,
	}

	category := addCategory(t, "test_category")
	mood := addMood(t, "test_mood")

	_, err := service.UploadNewSong(UploadNewSongRequest{
		Name:         "test_song",
		LyricAuthor:  null.StringFrom("test lyric author"),
		MusicAuthor:  "test music author",
		CategoryId:   category.Id,
		MoodId:       mood.Id,
		FullSongUrl:  "https://full-url.com",
		ShortSongUrl: "https://short-url.com",
		ImageUrl:     "https://image-url.com",
		Hashtags:     []string{"test"},
	}, gormDb, executionData)
	assert.Nil(t, err)

	var song database.CreatorSong
	if err = gormDb.Find(&song).Error; err != nil {
		t.Fatal(err)
	}

	assert.Greater(t, song.Id, int64(0))
	assert.Equal(t, song.UserId, userId)
	assert.Equal(t, song.Name, "test_song")
}
