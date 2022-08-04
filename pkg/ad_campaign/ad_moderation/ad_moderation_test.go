package ad_moderation

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/ads-manager/configs"
	converter2 "github.com/digitalmonsters/ads-manager/pkg/converter"
	"github.com/digitalmonsters/ads-manager/pkg/database"
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/digitalmonsters/go-common/wrappers"
	"github.com/digitalmonsters/go-common/wrappers/content"
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

var gormDb *gorm.DB
var adModerationService IService

func TestMain(m *testing.M) {
	gormDb = database.GetDb(database.DbTypeMaster)

	userWrapperMock := &user_go.UserGoWrapperMock{
		GetUsersFn: func(userIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]user_go.UserRecord] {
			ch := make(chan wrappers.GenericResponseChan[map[int64]user_go.UserRecord], 2)
			go func() {
				defer close(ch)
				var userMap = make(map[int64]user_go.UserRecord)
				for _, userId := range userIds {
					userMap[userId] = user_go.UserRecord{
						UserId:    userId,
						Username:  fmt.Sprintf("username_%v", userId),
						Firstname: fmt.Sprintf("first_%v", userId),
						Lastname:  fmt.Sprintf("last_%v", userId),
						Email:     fmt.Sprintf("email_%v", userId),
					}
				}
				ch <- wrappers.GenericResponseChan[map[int64]user_go.UserRecord]{
					Response: userMap,
				}
			}()
			return ch
		},
	}
	contentWrapper := &content.ContentWrapperMock{
		GetInternalFn: func(contentIds []int64, includeDeleted bool, apmTransaction *apm.Transaction,
			forceLog bool) chan wrappers.GenericResponseChan[map[int64]content.SimpleContent] {
			ch := make(chan wrappers.GenericResponseChan[map[int64]content.SimpleContent], 2)
			defer close(ch)
			var contentMap = make(map[int64]content.SimpleContent)
			for _, id := range contentIds {
				contentMap[id] = content.SimpleContent{
					Id:      id,
					VideoId: fmt.Sprint(id),
				}
			}
			ch <- wrappers.GenericResponseChan[map[int64]content.SimpleContent]{
				Response: contentMap,
			}
			return ch
		},
	}
	converter := converter2.NewConverter(userWrapperMock, contentWrapper)
	adModerationService = NewService(nil, converter)

	os.Exit(m.Run())
}

func TestService_GetAdModerationRequests(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresAllTables(configs.GetConfig().MasterDb, nil, nil); err != nil {
		t.Fatal(err)
	}
	configs.SetMockAppConfig(configs.AppConfig{ADS_MODERATION_SLA: 48})
	var addCampaigns = []database.AdCampaign{
		database.AdCampaign{
			UserId:         1,
			Name:           "test campaign 1",
			AdType:         database.AdTypeContent,
			Status:         database.AdCampaignStatusPending,
			ContentId:      123,
			Link:           null.StringFrom("test link 1"),
			LinkButtonId:   null.Int{},
			Country:        null.StringFrom("UK"),
			StartedAt:      null.TimeFrom(time.Now().UTC().Add(-24 * time.Hour)),
			EndedAt:        null.TimeFrom(time.Now().UTC().Add(24 * time.Hour)),
			DurationMin:    267,
			Budget:         decimal.NewFromInt(198),
			Gender:         null.StringFrom("male"),
			AgeFrom:        22,
			AgeTo:          34,
			RejectReasonId: null.Int{},
			CreatedAt:      time.Now().UTC().Add(-124 * time.Hour),
		},
		database.AdCampaign{
			UserId:         2,
			Name:           "test campaign 2",
			AdType:         database.AdTypeLink,
			Status:         database.AdCampaignStatusModerated,
			ContentId:      1213,
			Link:           null.StringFrom("test link 2"),
			LinkButtonId:   null.Int{},
			Country:        null.StringFrom("IT"),
			StartedAt:      null.TimeFrom(time.Now().UTC().Add(-5 * time.Hour)),
			EndedAt:        null.TimeFrom(time.Now().UTC().Add(256 * time.Hour)),
			DurationMin:    2178,
			Budget:         decimal.NewFromInt(1671),
			Gender:         null.StringFrom("female"),
			AgeFrom:        19,
			AgeTo:          25,
			RejectReasonId: null.Int{},
		},
		database.AdCampaign{
			UserId:         3,
			Name:           "test campaign 3",
			AdType:         database.AdTypeContent,
			Status:         database.AdCampaignStatusReject,
			ContentId:      127,
			Link:           null.StringFrom("test link 3"),
			LinkButtonId:   null.Int{},
			Country:        null.StringFrom("Spain"),
			StartedAt:      null.TimeFrom(time.Now().UTC().Add(-2 * time.Hour)),
			EndedAt:        null.TimeFrom(time.Now().UTC().Add(122 * time.Hour)),
			DurationMin:    22,
			Budget:         decimal.NewFromInt(18),
			Gender:         null.StringFrom("male"),
			AgeFrom:        34,
			AgeTo:          56,
			RejectReasonId: null.IntFrom(1),
		},
	}
	if err := gormDb.Create(&addCampaigns).Error; err != nil {
		t.Fatal(err)
	}

	resp, err := adModerationService.GetAdModerationRequests(GetAdModerationRequest{
		UserId:               null.Int{},
		Status:               nil,
		StartedAtFrom:        null.Time{},
		StartedAtTo:          null.Time{},
		EndedAtFrom:          null.Time{},
		EndedAtTo:            null.Time{},
		MaxThresholdExceeded: null.Bool{},
		Limit:                10,
		Offset:               0,
	}, gormDb, context.TODO())

	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 3, len(resp.Items))
	assert.Equal(t, int64(3), resp.TotalCount)

	resp, err = adModerationService.GetAdModerationRequests(GetAdModerationRequest{
		UserId:               null.IntFrom(1),
		Status:               []database.AdCampaignStatus{database.AdCampaignStatusPending},
		StartedAtFrom:        null.TimeFrom(time.Now().Add(-150 * time.Hour)),
		StartedAtTo:          null.TimeFrom(time.Now().Add(200 * time.Hour)),
		EndedAtFrom:          null.TimeFrom(time.Now().Add(-150 * time.Hour)),
		EndedAtTo:            null.TimeFrom(time.Now().Add(200 * time.Hour)),
		MaxThresholdExceeded: null.BoolFrom(true),
		Limit:                10,
		Offset:               0,
	}, gormDb, context.TODO())

	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(resp.Items))
	assert.True(t, len(resp.Items[0].Username) > 0)
	assert.Equal(t, int64(1), resp.TotalCount)
}

func TestService_SetAdRejectReason(t *testing.T) {

	cases := []struct {
		description string
		setStatus   database.AdCampaignStatus
	}{
		{
			description: "reject",
			setStatus:   database.AdCampaignStatusReject,
		},
		{
			description: "approved",
			setStatus:   database.AdCampaignStatusModerated,
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {

			if err := boilerplate_testing.FlushPostgresAllTables(configs.GetConfig().MasterDb, nil, nil); err != nil {
				t.Fatal(err)
			}
			configs.SetMockAppConfig(configs.AppConfig{ADS_MODERATION_SLA: 48})

			var addCampaign = database.AdCampaign{
				UserId:         1,
				Name:           "test campaign 1",
				AdType:         database.AdTypeContent,
				Status:         database.AdCampaignStatusPending,
				ContentId:      123,
				Link:           null.StringFrom("test link 1"),
				LinkButtonId:   null.Int{},
				Country:        null.StringFrom("UK"),
				StartedAt:      null.TimeFrom(time.Now().UTC().Add(-24 * time.Hour)),
				EndedAt:        null.TimeFrom(time.Now().UTC().Add(24 * time.Hour)),
				DurationMin:    267,
				Budget:         decimal.NewFromInt(198),
				Gender:         null.StringFrom("male"),
				AgeFrom:        22,
				AgeTo:          34,
				RejectReasonId: null.Int{},
				CreatedAt:      time.Now().UTC().Add(-124 * time.Hour),
			}
			if err := gormDb.Create(&addCampaign).Error; err != nil {
				t.Fatal(err)
			}
			var rejectReason = database.RejectReason{
				Reason:    "test reason",
				CreatedAt: time.Time{},
			}
			if err := gormDb.Create(&rejectReason).Error; err != nil {
				t.Fatal(err)
			}
			var req = SetAdRejectReasonRequest{
				Id:     addCampaign.Id,
				Status: c.setStatus,
			}
			if req.Status == database.AdCampaignStatusReject {
				req.RejectReasonId = null.IntFrom(rejectReason.Id)
			}

			resp, callbacks, err := adModerationService.SetAdRejectReason(req, gormDb, context.TODO())
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, 1, len(callbacks))
			assert.Equal(t, addCampaign.Id, resp.Id)
			assert.Equal(t, c.setStatus, resp.Status)
			if c.setStatus == database.AdCampaignStatusReject {
				assert.Equal(t, rejectReason.Id, resp.RejectReasonId.Int64)
			}
		})
	}
}
