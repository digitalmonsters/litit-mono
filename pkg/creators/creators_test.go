package creators

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/wrappers"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/digitalmonsters/go-common/wrappers/follow"
	"github.com/digitalmonsters/go-common/wrappers/music"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/music/configs"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/digitalmonsters/music/pkg/feed/feed_converter"
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

	feedConverter := feed_converter.NewFeedConverter(userWrapper, followWrapper, context.Background())

	service = NewService(feedConverter, nil)

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

	err := service.BecomeMusicCreator(BecomeMusicCreatorRequest{LibraryLink: link}, gormDb, executionData, userWrapper)
	assert.Nil(t, err)

	var creatorRequest database.Creator
	err = gormDb.Model(creatorRequest).Where("user_id = ?", userId).Find(&creatorRequest).Error
	assert.Nil(t, err)
	assert.Equal(t, creatorRequest.LibraryUrl, link)

	err = service.BecomeMusicCreator(BecomeMusicCreatorRequest{LibraryLink: link}, gormDb, executionData, userWrapper)
	assert.NotNil(t, err)
}

func TestCreatorRequestsList(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresTables(config.MasterDb, []string{"public.creators"}, nil, t); err != nil {
		t.Fatal(err)
	}

	var requests []database.Creator
	for i := 0; i < 10; i++ {
		r := database.Creator{
			Status:     user_go.CreatorStatusPending,
			UserId:     int64(i),
			LibraryUrl: fmt.Sprintf("https://music%v", i),
		}

		if i%2 == 0 {
			r.Status = user_go.CreatorStatusApproved
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
		Username:  "username100",
		Firstname: "firstname100",
		Lastname:  "lastname100",
		Email:     "email100",
		Status:    user_go.CreatorStatusPending,
		CreatedAt: time.Date(2020, 1, 1, 1, 1, 1, 1, time.UTC),
	}

	if err := gormDb.Create(&thresholdRequest).Error; err != nil {
		t.Fatal(err)
	}

	resp, err := service.CreatorRequestsList(CreatorRequestsListRequest{
		Statuses:             []user_go.CreatorStatus{user_go.CreatorStatusPending},
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

	resp, err = service.CreatorRequestsList(CreatorRequestsListRequest{
		MaxThresholdExceeded: true,
		Limit:                10,
		Offset:               0,
		SearchQuery:          "firstname",
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
			Status:     user_go.CreatorStatusPending,
			UserId:     int64(i),
			LibraryUrl: fmt.Sprintf("https://music%v", i),
		}

		if i%2 == 0 {
			r.Status = user_go.CreatorStatusApproved
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
		assert.Equal(t, r.Status, user_go.CreatorStatusApproved)
	}
}

func TestCreatorRequestReject(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresTables(config.MasterDb, []string{"public.creators", "public.creator_reject_reasons"}, nil, t); err != nil {
		t.Fatal(err)
	}

	var requests []database.Creator
	for i := 0; i < 10; i++ {
		r := database.Creator{
			Status:     user_go.CreatorStatusPending,
			UserId:     int64(i),
			LibraryUrl: fmt.Sprintf("https://music%v", i),
		}

		if i%2 == 0 {
			r.Status = user_go.CreatorStatusRejected
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
		assert.Equal(t, r.Status, user_go.CreatorStatusRejected)
	}
}

func addCreator(t *testing.T, userId int64, status user_go.CreatorStatus) *database.Creator {
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

	contentWrapper := &content.ContentWrapperMock{}

	contentWrapper.InsertMusicContentFn = func(content1 content.MusicContentRequest, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[content.SimpleContent] {
		ch := make(chan wrappers.GenericResponseChan[content.SimpleContent], 2)
		ch <- wrappers.GenericResponseChan[content.SimpleContent]{
			Error: nil,
			Response: content.SimpleContent{
				Id: 1,
			},
		}
		close(ch)

		return ch
	}

	userId := int64(111)
	creator := addCreator(t, userId, user_go.CreatorStatusApproved)
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
	}, contentWrapper, gormDb, executionData)
	assert.Nil(t, err)

	var song database.CreatorSong
	if err = gormDb.Find(&song).Error; err != nil {
		t.Fatal(err)
	}

	assert.Greater(t, song.Id, int64(0))
	assert.Equal(t, song.UserId, userId)
	assert.Equal(t, song.Name, "test_song")
}

func TestService_SongsList(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresTables(config.MasterDb, []string{"" +
		"public.creators", "public.categories", "public.creator_songs", "public.moods"}, nil, t); err != nil {
		t.Fatal(err)
	}

	category := addCategory(t, "test_category")
	mood := addMood(t, "test_mood")
	userId := int64(1)

	var songs []database.CreatorSong

	for i := 1; i <= 10; i++ {
		s := database.CreatorSong{
			Name:       fmt.Sprintf("test song %v", i),
			Status:     music.CreatorSongStatusApproved,
			CategoryId: category.Id,
			MoodId:     mood.Id,
		}

		if i%2 == 0 {
			s.UserId = userId
		} else {
			s.UserId = 2
		}

		songs = append(songs, s)
	}

	if err := gormDb.Create(&songs).Error; err != nil {
		t.Fatal(err)
	}

	resp, err := service.SongsList(SongsListRequest{
		UserId: 1,
		Count:  10,
	}, 0, gormDb, router.MethodExecutionData{
		Context: context.Background(),
	})

	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, resp.Items, 5)
	for _, i := range resp.Items {
		assert.Equal(t, i.UserId, userId)
	}
}
