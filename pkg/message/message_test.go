package message

import (
	"github.com/digitalmonsters/ads-manager/configs"
	"github.com/digitalmonsters/ads-manager/pkg/database"
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/shopspring/decimal"
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
				Title:              "test1",
				Description:        "desc_test_1",
				Countries:          []string{"UA"},
				IsActive:           true,
				VerificationStatus: database.VerificationStatusVerified,
			},
			{
				Title:              "test2",
				Description:        "desc_test_2",
				Countries:          []string{"UA", "US"},
				VerificationStatus: database.VerificationStatusPending,
			},
			{
				Title:              "test3",
				Description:        "desc_test_3",
				Countries:          []string{"RU"},
				VerificationStatus: database.VerificationStatusRejected,
			},
		},
	}, gormDb)
	assert.Nil(t, err)
	assert.Len(t, resp, 3)

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
	assert.Len(t, list.Items, 2)

	list, err = MessagesListAdmin(MessagesListAdminRequest{
		Limit:     10,
		Offset:    0,
		Countries: []string{"RU"},
	}, gormDb)
	assert.Nil(t, err)
	assert.Len(t, list.Items, 1)
}

func TestGetMessageForUser(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresTables(config.MasterDb, []string{"public.messages"}, nil, t); err != nil {
		t.Fatal(err)
	}

	resp, err := UpsertMessageBulkAdmin(UpsertMessageAdminRequest{
		Items: []adminMessage{
			{
				Title:              "test1",
				Description:        "desc_test_1",
				Countries:          []string{"UA"},
				AgeFrom:            18,
				AgeTo:              25,
				IsActive:           true,
				VerificationStatus: database.VerificationStatusVerified,
			},
		},
	}, gormDb)
	assert.Nil(t, err)
	assert.Len(t, resp, 1)

	userId := int64(1)
	userGoWrapper.GetUsersDetailFn = func(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan user_go.GetUsersDetailsResponseChan {
		respCh := make(chan user_go.GetUsersDetailsResponseChan, 2)
		defer close(respCh)

		result := user_go.GetUsersDetailsResponseChan{
			Error: nil,
			Items: map[int64]user_go.UserDetailRecord{
				userId: {
					Id:          userId,
					Firstname:   "test",
					Lastname:    "test",
					Birthdate:   null.TimeFrom(time.Date(2002, 1, 1, 1, 1, 1, 1, time.UTC)),
					KycStatus:   "verified",
					CountryCode: "UA",
					VaultPoints: decimal.NewFromFloat(200),
				},
			},
		}

		respCh <- result

		return respCh
	}

	message, err := GetMessageForUser(userId, gormDb, userGoWrapper, nil)
	assert.Nil(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, message.Title, resp[0].Title)
	assert.Equal(t, message.Description, resp[0].Description)
}
