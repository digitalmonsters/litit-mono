package message

import (
	"github.com/digitalmonsters/ads-manager/configs"
	"github.com/digitalmonsters/ads-manager/pkg/database"
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/wrappers"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"os"
	"testing"
	"time"
)

var config configs.Settings
var gormDb *gorm.DB
var userGoWrapper *user_go.UserGoWrapperMock

func TestMain(m *testing.M) {
	config = configs.GetConfig()
	gormDb = database.GetDb(database.DbTypeMaster)
	userGoWrapper = &user_go.UserGoWrapperMock{}
	os.Exit(m.Run())
}

func TestMethods(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresTables(config.MasterDb, []string{"public.messages"}, nil, t); err != nil {
		t.Fatal(err)
	}

	resp, err := UpsertMessageBulkAdmin(UpsertMessageAdminRequest{
		Items: []adminMessage{
			{
				Type:               database.MessageTypeMobile,
				Title:              "test1",
				Description:        "desc_test_1",
				Countries:          []string{"UA"},
				IsActive:           true,
				VerificationStatus: null.IntFrom(int64(database.VerificationStatusVerified)),
			},
			{
				Type:               database.MessageTypeMobile,
				Title:              "test2",
				Description:        "desc_test_2",
				Countries:          []string{"UA", "US"},
				IsActive:           true,
				VerificationStatus: null.IntFrom(int64(database.VerificationStatusPending)),
			},
			{
				Type:        database.MessageTypeMobile,
				Title:       "test3",
				Description: "desc_test_3",
				IsActive:    true,
				Countries:   []string{"RU"},
			},
			{
				Type:        database.MessageTypeMobile,
				Title:       "test4",
				Description: "desc_test_4",
				PointsFrom:  1,
				PointsTo:    10,
				IsActive:    true,
				Countries:   []string{"UK"},
			},
			{
				Type:        database.MessageTypeMobile,
				Title:       "test5",
				Description: "desc_test_5",
				PointsFrom:  1,
				PointsTo:    2,
				IsActive:    false,
				Countries:   []string{"RU"},
			},
		},
	}, gormDb)
	assert.Nil(t, err)
	assert.Len(t, resp, 5)

	resp, err = UpsertMessageBulkAdmin(UpsertMessageAdminRequest{
		Items: []adminMessage{
			{
				Id:       null.IntFrom(resp[0].Id),
				AgeFrom:  10,
				AgeTo:    20,
				IsActive: false,
			},
		},
	}, gormDb)
	assert.Nil(t, err)

	var record database.Message
	err = gormDb.Where("id = ?", resp[0].Id).Find(&record).Error
	assert.Nil(t, err)
	assert.Equal(t, record.AgeFrom, int8(10))
	assert.NotNil(t, record.DeactivatedAt)

	err = DeleteMessagesBulkAdmin(DeleteMessagesBulkAdminRequest{
		Ids: []int64{record.Id},
	}, gormDb)
	assert.Nil(t, err)

	err = gormDb.Where("id = ?", record.Id).First(&record).Error
	assert.NotNil(t, err)

	list, err := MessagesListAdmin(MessagesListAdminRequest{
		Limit:  10,
		Offset: 0,
	}, gormDb)
	assert.Nil(t, err)
	assert.Len(t, list.Items, 4)

	list, err = MessagesListAdmin(MessagesListAdminRequest{
		Limit:     10,
		Offset:    0,
		Countries: []string{"RU"},
	}, gormDb)
	assert.Nil(t, err)
	assert.Len(t, list.Items, 2)

	list1, err1 := MessagesListAdmin(MessagesListAdminRequest{
		Limit:    10,
		Offset:   0,
		IsActive: null.BoolFrom(true),
	}, gormDb)
	assert.Nil(t, err1)
	assert.Equal(t, 3, len(list1.Items))

	list2, err1 := MessagesListAdmin(MessagesListAdminRequest{
		Limit:              10,
		Offset:             0,
		IsActive:           null.BoolFrom(false),
		VerificationStatus: nil,
	}, gormDb)
	assert.Nil(t, err1)
	assert.Equal(t, 1, len(list2.Items))

	list3, err3 := MessagesListAdmin(MessagesListAdminRequest{
		Limit:              10,
		Offset:             0,
		IsActive:           null.BoolFrom(true),
		Countries:          []string{"RU", "US"},
		VerificationStatus: nil,
	}, gormDb)
	assert.Nil(t, err3)
	assert.Equal(t, 2, len(list3.Items))

	list4, err4 := MessagesListAdmin(MessagesListAdminRequest{
		Limit:              10,
		Offset:             0,
		IsActive:           null.BoolFrom(true),
		Countries:          []string{"RU", "UK"},
		VerificationStatus: nil,
	}, gormDb)
	assert.Nil(t, err4)
	assert.Equal(t, 2, len(list4.Items))

	list5, err5 := MessagesListAdmin(MessagesListAdminRequest{
		Limit:              10,
		Offset:             0,
		PointsToFrom:       2,
		PointsToTo:         10,
		VerificationStatus: nil,
	}, gormDb)
	assert.Nil(t, err5)
	assert.Equal(t, 2, len(list5.Items))

	list6, err6 := MessagesListAdmin(MessagesListAdminRequest{
		Limit:              10,
		Offset:             0,
		PointsFromFrom:     1,
		PointsFromTo:       10,
		VerificationStatus: nil,
	}, gormDb)
	assert.Nil(t, err6)
	assert.Equal(t, 2, len(list6.Items))
}

func TestGetMessageForUser(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresTables(config.MasterDb, []string{"public.messages"}, nil, t); err != nil {
		t.Fatal(err)
	}

	statusVerified := database.VerificationStatusVerified

	resp, err := UpsertMessageBulkAdmin(UpsertMessageAdminRequest{
		Items: []adminMessage{
			{
				Type:        database.MessageTypeMobile,
				Title:       "test1",
				Description: "desc_test_1",
				Countries:   []string{"UA"},
				AgeFrom:     18,
				AgeTo:       25,
				IsActive:    true,
			},
			{
				Type:               database.MessageTypeMobile,
				Title:              "test2",
				Description:        "desc_test_2",
				Countries:          []string{"UA"},
				AgeFrom:            18,
				AgeTo:              25,
				IsActive:           true,
				VerificationStatus: null.IntFrom(int64(statusVerified)),
			},
			{
				Type:        database.MessageTypeWeb,
				Title:       "front_test",
				Description: "front_test_d",
				IsActive:    true,
			},
		},
	}, gormDb)
	assert.Nil(t, err)
	assert.Len(t, resp, 3)

	userId := int64(1)

	userGoWrapper.GetUsersDetailFn = func(userIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]user_go.UserDetailRecord] {
		ch := make(chan wrappers.GenericResponseChan[map[int64]user_go.UserDetailRecord], 2)
		defer close(ch)

		userMap := map[int64]user_go.UserDetailRecord{}
		for _, id := range userIds {
			userMap[id] = user_go.UserDetailRecord{
				Id:          userId,
				Firstname:   "test",
				Lastname:    "test",
				Birthdate:   null.TimeFrom(time.Date(2002, 1, 1, 1, 1, 1, 1, time.UTC)),
				KycStatus:   "verified",
				CountryCode: "UA",
				VaultPoints: decimal.NewFromFloat(200),
			}
		}

		ch <- wrappers.GenericResponseChan[map[int64]user_go.UserDetailRecord]{
			Error:    nil,
			Response: userMap,
		}

		return ch
	}

	message, err := GetMessageForUser(userId, database.MessageTypeMobile, gormDb, userGoWrapper, router.MethodExecutionData{
		Context: context.TODO(),
	})
	assert.Nil(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, message.Title, resp[1].Title)
	assert.Equal(t, message.Description, resp[1].Description)

	message, err = GetMessageForUser(userId, database.MessageTypeWeb, gormDb, userGoWrapper, router.MethodExecutionData{
		Context: context.TODO(),
	})
	assert.Nil(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, message.Title, resp[2].Title)
	assert.Equal(t, message.Description, resp[2].Description)

}
