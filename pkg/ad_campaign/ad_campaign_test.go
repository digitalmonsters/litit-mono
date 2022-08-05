package ad_campaign

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/ads-manager/configs"
	"github.com/digitalmonsters/ads-manager/pkg/database"
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/digitalmonsters/go-common/filters"
	"github.com/digitalmonsters/go-common/wrappers"
	"github.com/digitalmonsters/go-common/wrappers/ads_manager"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/digitalmonsters/go-common/wrappers/go_tokenomics"
	"github.com/digitalmonsters/go-common/wrappers/user_category"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"os"
	"strings"
	"testing"
	"time"
)

var gormDb *gorm.DB
var adCampaignService IService
var contentWrapperMock content.IContentWrapper
var userCategoryWrapper *user_category.UserCategoryWrapperMock
var userWrapper user_go.IUserGoWrapper
var goTokenomicsWrapper *go_tokenomics.GoTokenomicsWrapperMock

func TestMain(m *testing.M) {
	gormDb = database.GetDb(database.DbTypeMaster)

	contentWrapperMock = &content.ContentWrapperMock{
		GetInternalFn: func(contentIds []int64, includeDeleted bool, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[int64]content.SimpleContent] {
			ch := make(chan wrappers.GenericResponseChan[map[int64]content.SimpleContent], 2)
			go func() {
				defer func() {
					close(ch)
				}()

				contentsMap := make(map[int64]content.SimpleContent, len(contentIds))
				for _, contentId := range contentIds {
					contentsMap[contentId] = content.SimpleContent{Id: contentId}
				}
				ch <- wrappers.GenericResponseChan[map[int64]content.SimpleContent]{
					Error:    nil,
					Response: contentsMap,
				}
			}()

			return ch
		},
		GetCategoryInternalFn: func(categoryIds []int64, omitCategoryIds []int64, limit int, offset int,
			onlyParent null.Bool, withViews null.Bool, apmTransaction *apm.Transaction, shouldHaveValidContent bool,
			forceLog bool) chan wrappers.GenericResponseChan[content.CategoryResponseData] {
			ch := make(chan wrappers.GenericResponseChan[content.CategoryResponseData], 2)

			go func() {
				defer func() {
					close(ch)
				}()

				categories := make([]content.SimpleCategoryModel, len(categoryIds))
				for i, categoryId := range categoryIds {
					categories[i] = content.SimpleCategoryModel{
						Id:         categoryId,
						Name:       fmt.Sprintf("category_%v", categoryId),
						ViewsCount: 100,
						Emojis:     "",
					}
				}
				ch <- wrappers.GenericResponseChan[content.CategoryResponseData]{
					Error: nil,
					Response: content.CategoryResponseData{
						Items:      categories,
						TotalCount: int64(len(categories)),
					},
				}
			}()

			return ch
		},
	}

	userCategoryWrapper = &user_category.UserCategoryWrapperMock{
		GetInternalUserCategorySubscriptionsFn: func(userId int64, limit int, pageState string, ctx context.Context,
			forceLog bool) chan wrappers.GenericResponseChan[user_category.GetInternalUserCategorySubscriptionsResponse] {
			ch := make(chan wrappers.GenericResponseChan[user_category.GetInternalUserCategorySubscriptionsResponse], 2)
			go func() {
				defer func() {
					close(ch)
				}()

				categoryIds := make([]int64, limit)
				for i := 0; i < limit; i++ {
					categoryIds[i] = int64(i + 1)
				}

				ch <- wrappers.GenericResponseChan[user_category.GetInternalUserCategorySubscriptionsResponse]{
					Error: nil,
					Response: user_category.GetInternalUserCategorySubscriptionsResponse{
						CategoryIds: categoryIds,
						PageState:   "",
					},
				}
			}()

			return ch
		},
	}

	userWrapper = &user_go.UserGoWrapperMock{
		GetUserDetailsFn: func(userId int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[user_go.UserDetailRecord] {
			var respCh = make(chan wrappers.GenericResponseChan[user_go.UserDetailRecord], 1)
			respCh <- wrappers.GenericResponseChan[user_go.UserDetailRecord]{
				Error: nil,
				Response: user_go.UserDetailRecord{
					Id:          userId,
					CountryCode: "US",
					Gender:      null.StringFrom("male"),
					AdDisabled:  false,
					Birthdate:   null.TimeFrom(time.Now().UTC().Add(-time.Hour * 24 * 365 * 30)),
				},
			}
			close(respCh)
			return respCh
		},
	}

	goTokenomicsWrapper = &go_tokenomics.GoTokenomicsWrapperMock{}

	goTokenomicsWrapper.GetUsersTokenomicsInfoFn = func(userIds []int64, filters []filters.Filter, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]go_tokenomics.UserTokenomicsInfo] {
		var respCh = make(chan wrappers.GenericResponseChan[map[int64]go_tokenomics.UserTokenomicsInfo], 1)
		resp := make(map[int64]go_tokenomics.UserTokenomicsInfo, len(userIds))
		for _, userId := range userIds {
			resp[userId] = go_tokenomics.UserTokenomicsInfo{
				TotalPoints:        decimal.NewFromInt(100),
				CurrentPoints:      decimal.NewFromInt(100),
				VaultPoints:        decimal.NewFromInt(100),
				AllTimeVaultPoints: decimal.NewFromInt(100),
				CurrentTokens:      decimal.NewFromInt(100),
				CurrentRate:        decimal.NewFromInt(1),
				WithdrawnTokens:    decimal.NewFromInt(100),
			}
		}

		respCh <- wrappers.GenericResponseChan[map[int64]go_tokenomics.UserTokenomicsInfo]{
			Error:    nil,
			Response: resp,
		}
		close(respCh)
		return respCh
	}

	adCampaignService = NewService(contentWrapperMock, userCategoryWrapper, userWrapper, nil, goTokenomicsWrapper)

	os.Exit(m.Run())
}

func TestService_CreateAdCampaign(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresAllTables(configs.GetConfig().MasterDb, nil, nil); err != nil {
		t.Fatal(err)
	}

	configs.SetMockAppConfig(configs.AppConfig{
		ADS_CAMPAIGN_GLOBAL_PRICE: decimal.NewFromInt(10),
	})

	tx := gormDb.Begin()
	defer tx.Rollback()

	if err := adCampaignService.CreateAdCampaign(CreateAdCampaignRequest{
		Name:          "a1",
		AdType:        database.AdTypeContent,
		ContentId:     1,
		Link:          null.StringFrom("l1"),
		LinkButtonId:  null.IntFrom(1),
		Country:       null.StringFrom("usa"),
		DurationMin:   1,
		Budget:        decimal.NewFromInt(1),
		Gender:        null.StringFrom("male"),
		AgeFrom:       0,
		AgeTo:         100,
		CategoriesIds: []int64{1, 2, 3},
	}, 1, tx, context.TODO()); err != nil {
		t.Fatal(err)
	}

	if err := tx.Commit().Error; err != nil {
		t.Fatal(err)
	}

	var adCampaign database.AdCampaign
	if err := gormDb.First(&adCampaign).Error; err != nil {
		t.Fatal(err)
	}

	a := assert.New(t)

	a.True(strings.EqualFold("a1", adCampaign.Name))
}

func TestService_GetAdsContentForUser(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresAllTables(configs.GetConfig().MasterDb, nil, nil); err != nil {
		t.Fatal(err)
	}

	configs.SetMockAppConfig(configs.AppConfig{
		ADS_CAMPAIGN_VIDEOS_PER_CONTENT_VIDEOS: 2,
	})

	if err := gormDb.Create(&database.AdCampaign{
		Id:          1,
		UserId:      1,
		Name:        "ad1",
		AdType:      database.AdTypeContent,
		Status:      database.AdCampaignStatusActive,
		ContentId:   7,
		DurationMin: 10,
		Budget:      decimal.NewFromInt(10),
		Gender:      null.StringFrom("male"),
		AgeFrom:     20,
		AgeTo:       50,
	}).Error; err != nil {
		t.Fatal(err)
	}

	if err := gormDb.Create(&database.AdCampaign{
		Id:          2,
		UserId:      1,
		Name:        "ad2",
		AdType:      database.AdTypeContent,
		Status:      database.AdCampaignStatusActive,
		ContentId:   9,
		DurationMin: 10,
		Budget:      decimal.NewFromInt(10),
		Gender:      null.StringFrom("male"),
		AgeFrom:     20,
		AgeTo:       50,
	}).Error; err != nil {
		t.Fatal(err)
	}

	if err := gormDb.Create(&database.ActionButton{
		Id:   1,
		Name: "link_button1",
		Type: database.LinkButtonType,
	}).Error; err != nil {
		t.Fatal(err)
	}

	if err := gormDb.Create(&database.AdCampaign{
		Id:           3,
		UserId:       1,
		Name:         "ad3",
		AdType:       database.AdTypeLink,
		Status:       database.AdCampaignStatusActive,
		Link:         null.StringFrom("link1"),
		LinkButtonId: null.IntFrom(1),
		ContentId:    10,
		DurationMin:  10,
		Budget:       decimal.NewFromInt(10),
		Gender:       null.StringFrom("male"),
		AgeFrom:      20,
		AgeTo:        50,
	}).Error; err != nil {
		t.Fatal(err)
	}

	if err := gormDb.Create(&database.AdCampaignCategory{
		AdCampaignId: 3,
		CategoryId:   1,
		CategoryName: "category_1",
	}).Error; err != nil {
		t.Fatal(err)
	}

	if err := gormDb.Create(&database.AdCampaignCategory{
		AdCampaignId: 3,
		CategoryId:   2,
		CategoryName: "category_2",
	}).Error; err != nil {
		t.Fatal(err)
	}

	resp, err := adCampaignService.GetAdsContentForUser(ads_manager.GetAdsContentForUserRequest{
		UserId:             1,
		ContentIdsToMix:    []int64{1, 2, 3, 4, 5, 6},
		ContentIdsToIgnore: []int64{7, 8},
	}, gormDb, context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	a := assert.New(t)

	a.NotNil(resp)

	a.Len(resp.MixedContentIdsWithAd, 8)
	a.Equal(int64(1), resp.MixedContentIdsWithAd[0])
	a.Equal(int64(2), resp.MixedContentIdsWithAd[1])
	a.Equal(int64(9), resp.MixedContentIdsWithAd[2]) // ad
	a.Equal(int64(3), resp.MixedContentIdsWithAd[3])
	a.Equal(int64(4), resp.MixedContentIdsWithAd[4])
	a.Equal(int64(10), resp.MixedContentIdsWithAd[5]) // ad
	a.Equal(int64(5), resp.MixedContentIdsWithAd[6])
	a.Equal(int64(6), resp.MixedContentIdsWithAd[7])

	a.Len(resp.ContentAds, 2)
	contentAd, ok := resp.ContentAds[9]
	a.True(ok)
	a.False(contentAd.Link.Valid)
	a.False(contentAd.LinkButtonId.Valid)
	a.False(contentAd.LinkButtonName.Valid)
	contentAd, ok = resp.ContentAds[10]
	a.True(ok)
	a.True(contentAd.Link.Valid)
	a.True(strings.EqualFold("link1", contentAd.Link.String))
	a.True(contentAd.LinkButtonId.Valid)
	a.Equal(int64(1), contentAd.LinkButtonId.Int64)
	a.True(contentAd.LinkButtonName.Valid)
	a.True(strings.EqualFold("link_button1", contentAd.LinkButtonName.String))

	if err = gormDb.Create(&database.AdCampaign{
		Id:          4,
		UserId:      1,
		Name:        "ad4",
		AdType:      database.AdTypeContent,
		Status:      database.AdCampaignStatusActive,
		ContentId:   20,
		DurationMin: 10,
		Budget:      decimal.NewFromInt(10),
		Gender:      null.StringFrom("male"),
		AgeFrom:     20,
		AgeTo:       50,
	}).Error; err != nil {
		t.Fatal(err)
	}

	configs.SetMockAppConfig(configs.AppConfig{
		ADS_CAMPAIGN_VIDEOS_PER_CONTENT_VIDEOS: 9,
	})

	resp, err = adCampaignService.GetAdsContentForUser(ads_manager.GetAdsContentForUserRequest{
		UserId:             2,
		ContentIdsToMix:    []int64{1, 2, 3, 4, 5, 6, 7, 9, 10, 11},
		ContentIdsToIgnore: []int64{7, 8},
	}, gormDb, context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	a.NotNil(resp)

	a.Len(resp.MixedContentIdsWithAd, 11)
	a.Equal(int64(1), resp.MixedContentIdsWithAd[0])
	a.Equal(int64(2), resp.MixedContentIdsWithAd[1])
	a.Equal(int64(3), resp.MixedContentIdsWithAd[2])
	a.Equal(int64(4), resp.MixedContentIdsWithAd[3])
	a.Equal(int64(5), resp.MixedContentIdsWithAd[4])
	a.Equal(int64(6), resp.MixedContentIdsWithAd[5])
	a.Equal(int64(7), resp.MixedContentIdsWithAd[6])
	a.Equal(int64(9), resp.MixedContentIdsWithAd[7])
	a.Equal(int64(10), resp.MixedContentIdsWithAd[8]) // ad
	a.Equal(int64(20), resp.MixedContentIdsWithAd[9])
	a.Equal(int64(11), resp.MixedContentIdsWithAd[10])

	if err = gormDb.Create(&database.AdCampaign{
		Id:             5,
		UserId:         1,
		Name:           "ad5",
		AdType:         database.AdTypeContent,
		Status:         database.AdCampaignStatusActive,
		ContentId:      21,
		DurationMin:    15,
		Budget:         decimal.NewFromInt(1),
		OriginalBudget: decimal.NewFromInt(1),
		StartedAt:      null.TimeFrom(time.Now().UTC().Add(-5 * time.Minute)),
		EndedAt:        null.TimeFrom(time.Now().UTC().Add(10 * time.Minute)),
		Link:           null.StringFrom("link"),
		LinkButtonId:   null.IntFrom(0),
	}).Error; err != nil {
		t.Fatal(err)
	}

	resp, err = adCampaignService.GetAdsContentForUser(ads_manager.GetAdsContentForUserRequest{
		UserId:             2,
		ContentIdsToMix:    []int64{1, 2, 3, 4, 5, 6, 7, 9, 10, 11},
		ContentIdsToIgnore: []int64{7, 8, 20},
	}, gormDb, context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	a.NotNil(resp)

	a.Len(resp.MixedContentIdsWithAd, 11)
	a.Equal(int64(1), resp.MixedContentIdsWithAd[0])
	a.Equal(int64(2), resp.MixedContentIdsWithAd[1])
	a.Equal(int64(3), resp.MixedContentIdsWithAd[2])
	a.Equal(int64(4), resp.MixedContentIdsWithAd[3])
	a.Equal(int64(5), resp.MixedContentIdsWithAd[4])
	a.Equal(int64(6), resp.MixedContentIdsWithAd[5])
	a.Equal(int64(7), resp.MixedContentIdsWithAd[6])
	a.Equal(int64(9), resp.MixedContentIdsWithAd[7])
	a.Equal(int64(10), resp.MixedContentIdsWithAd[8]) // ad
	a.Equal(int64(21), resp.MixedContentIdsWithAd[9])
	a.Equal(int64(11), resp.MixedContentIdsWithAd[10])

	userCategoryWrapper.GetInternalUserCategorySubscriptionsFn = func(userId int64, limit int, pageState string, ctx context.Context,
		forceLog bool) chan wrappers.GenericResponseChan[user_category.GetInternalUserCategorySubscriptionsResponse] {
		ch := make(chan wrappers.GenericResponseChan[user_category.GetInternalUserCategorySubscriptionsResponse], 2)
		ch <- wrappers.GenericResponseChan[user_category.GetInternalUserCategorySubscriptionsResponse]{
			Error: nil,
			Response: user_category.GetInternalUserCategorySubscriptionsResponse{
				CategoryIds: nil,
				PageState:   "",
			},
		}
		close(ch)
		return ch
	}

	resp, err = adCampaignService.GetAdsContentForUser(ads_manager.GetAdsContentForUserRequest{
		UserId:             2,
		ContentIdsToMix:    []int64{1, 2, 3, 4, 5, 6, 7, 9, 10, 11},
		ContentIdsToIgnore: []int64{7, 8, 20},
	}, gormDb, context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	a.NotNil(resp)

	a.Len(resp.MixedContentIdsWithAd, 11)
	a.Equal(int64(1), resp.MixedContentIdsWithAd[0])
	a.Equal(int64(2), resp.MixedContentIdsWithAd[1])
	a.Equal(int64(3), resp.MixedContentIdsWithAd[2])
	a.Equal(int64(4), resp.MixedContentIdsWithAd[3])
	a.Equal(int64(5), resp.MixedContentIdsWithAd[4])
	a.Equal(int64(6), resp.MixedContentIdsWithAd[5])
	a.Equal(int64(7), resp.MixedContentIdsWithAd[6])
	a.Equal(int64(9), resp.MixedContentIdsWithAd[7])
	a.Equal(int64(10), resp.MixedContentIdsWithAd[8]) // ad
	a.Equal(int64(21), resp.MixedContentIdsWithAd[9])
	a.Equal(int64(11), resp.MixedContentIdsWithAd[10])

	if err = gormDb.Create(&database.AdCampaign{
		Id:             6,
		UserId:         1,
		Name:           "ad6",
		Status:         database.AdCampaignStatusActive,
		ContentId:      8,
		DurationMin:    15,
		Budget:         decimal.NewFromInt(1),
		OriginalBudget: decimal.NewFromInt(1),
		AgeFrom:        0,
		AgeTo:          100,
	}).Error; err != nil {
		t.Fatal(err)
	}

	if err = gormDb.Create(&database.AdCampaignView{
		AdCampaignId: 6,
		UserId:       2,
	}).Error; err != nil {
		t.Fatal(err)
	}

	resp, err = adCampaignService.GetAdsContentForUser(ads_manager.GetAdsContentForUserRequest{
		UserId:             2,
		ContentIdsToMix:    []int64{1, 2, 3, 4, 5, 6, 7, 9, 10, 11},
		ContentIdsToIgnore: []int64{7, 20},
	}, gormDb, context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	a.NotNil(resp)

	a.Len(resp.MixedContentIdsWithAd, 11)
	a.Equal(int64(1), resp.MixedContentIdsWithAd[0])
	a.Equal(int64(2), resp.MixedContentIdsWithAd[1])
	a.Equal(int64(3), resp.MixedContentIdsWithAd[2])
	a.Equal(int64(4), resp.MixedContentIdsWithAd[3])
	a.Equal(int64(5), resp.MixedContentIdsWithAd[4])
	a.Equal(int64(6), resp.MixedContentIdsWithAd[5])
	a.Equal(int64(7), resp.MixedContentIdsWithAd[6])
	a.Equal(int64(9), resp.MixedContentIdsWithAd[7])
	a.Equal(int64(10), resp.MixedContentIdsWithAd[8]) // ad
	a.Equal(int64(21), resp.MixedContentIdsWithAd[9])
	a.Equal(int64(11), resp.MixedContentIdsWithAd[10])

	configs.SetMockAppConfig(configs.AppConfig{
		ADS_CAMPAIGN_VIDEOS_PER_CONTENT_VIDEOS: 5,
	})

	resp, err = adCampaignService.GetAdsContentForUser(ads_manager.GetAdsContentForUserRequest{
		UserId:             2,
		ContentIdsToMix:    []int64{1, 2, 3, 4, 5, 6, 7, 9, 10, 11},
		ContentIdsToIgnore: []int64{7, 8, 21},
	}, gormDb, context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	a.NotNil(resp)

	a.Len(resp.MixedContentIdsWithAd, 11)
	a.Equal(int64(1), resp.MixedContentIdsWithAd[0])
	a.Equal(int64(2), resp.MixedContentIdsWithAd[1])
	a.Equal(int64(3), resp.MixedContentIdsWithAd[2])
	a.Equal(int64(4), resp.MixedContentIdsWithAd[3])
	a.Equal(int64(5), resp.MixedContentIdsWithAd[4])
	a.Equal(int64(20), resp.MixedContentIdsWithAd[5])
	a.Equal(int64(6), resp.MixedContentIdsWithAd[6])
	a.Equal(int64(7), resp.MixedContentIdsWithAd[7])
	a.Equal(int64(9), resp.MixedContentIdsWithAd[8])
	a.Equal(int64(10), resp.MixedContentIdsWithAd[9])
	a.Equal(int64(11), resp.MixedContentIdsWithAd[10])
}

func TestService_ClickLink(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresAllTables(configs.GetConfig().MasterDb, nil, nil); err != nil {
		t.Fatal(err)
	}

	if err := gormDb.Create(&database.AdCampaign{
		Id:        1,
		UserId:    1,
		Name:      "1",
		AdType:    database.AdTypeContent,
		Status:    database.AdCampaignStatusActive,
		ContentId: 1,
		Country:   null.StringFrom("us"),
		Budget:    decimal.NewFromInt(30),
		Price:     decimal.NewFromInt(10),
	}).Error; err != nil {
		t.Fatal(err)
	}

	if err := adCampaignService.ClickLink(2, ClickLinkRequest{ContentId: 1}, gormDb); err != nil {
		t.Fatal(err)
	}

	a := assert.New(t)

	var adCampaignClick database.AdCampaignClick
	if err := gormDb.Where("ad_campaign_id = 1 and user_id = 2").Find(&adCampaignClick).Error; err != nil {
		t.Fatal(err)
	}

	a.Equal(int64(1), adCampaignClick.AdCampaignId)
}

func TestService_StopAdCampaign(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresAllTables(configs.GetConfig().MasterDb, nil, nil); err != nil {
		t.Fatal(err)
	}

	if err := gormDb.Create(&database.AdCampaign{
		Id:        1,
		UserId:    1,
		Name:      "1",
		AdType:    database.AdTypeContent,
		Status:    database.AdCampaignStatusActive,
		ContentId: 1,
		Country:   null.StringFrom("us"),
		Budget:    decimal.NewFromInt(30),
		Price:     decimal.NewFromInt(10),
	}).Error; err != nil {
		t.Fatal(err)
	}

	if err := adCampaignService.StopAdCampaign(1, StopAdCampaignRequest{AdCampaignId: 1}, gormDb); err != nil {
		t.Fatal(err)
	}

	var adCampaign database.AdCampaign
	if err := gormDb.Where("id = 1").Find(&adCampaign).Error; err != nil {
		t.Fatal(err)
	}

	a := assert.New(t)
	a.Equal(database.AdCampaignStatusCompleted, adCampaign.Status)
}

func TestService_StartAdCampaign(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresAllTables(configs.GetConfig().MasterDb, nil, nil); err != nil {
		t.Fatal(err)
	}

	if err := gormDb.Create(&database.AdCampaign{
		Id:        1,
		UserId:    1,
		Name:      "1",
		AdType:    database.AdTypeContent,
		Status:    database.AdCampaignStatusModerated,
		ContentId: 1,
		Country:   null.StringFrom("us"),
		Budget:    decimal.NewFromInt(30),
		Price:     decimal.NewFromInt(10),
	}).Error; err != nil {
		t.Fatal(err)
	}

	if err := adCampaignService.StartAdCampaign(1, StartAdCampaignRequest{AdCampaignId: 1}, gormDb, context.TODO()); err != nil {
		t.Fatal(err)
	}

	var adCampaign database.AdCampaign
	if err := gormDb.Where("id = 1").Find(&adCampaign).Error; err != nil {
		t.Fatal(err)
	}

	a := assert.New(t)
	a.Equal(database.AdCampaignStatusActive, adCampaign.Status)
}

func TestService_ListAdCampaigns(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresAllTables(configs.GetConfig().MasterDb, nil, nil); err != nil {
		t.Fatal(err)
	}

	if err := gormDb.Create(&database.AdCampaign{
		Id:        1,
		UserId:    1,
		Name:      "ad1",
		AdType:    database.AdTypeContent,
		Status:    database.AdCampaignStatusModerated,
		ContentId: 1,
		Country:   null.StringFrom("us"),
		Budget:    decimal.NewFromInt(30),
		Price:     decimal.NewFromInt(10),
		CreatedAt: time.Now().UTC().Add(-1 * time.Hour),
	}).Error; err != nil {
		t.Fatal(err)
	}

	if err := gormDb.Create(&database.AdCampaign{
		Id:        2,
		UserId:    1,
		Name:      "ad2",
		AdType:    database.AdTypeContent,
		Status:    database.AdCampaignStatusActive,
		ContentId: 2,
		Country:   null.StringFrom("us"),
		Budget:    decimal.NewFromInt(30),
		Price:     decimal.NewFromInt(10),
		Views:     2,
		Clicks:    1,
		CreatedAt: time.Now().UTC().Add(-2 * time.Hour),
	}).Error; err != nil {
		t.Fatal(err)
	}

	if err := gormDb.Create(&database.AdCampaignView{
		AdCampaignId: 2,
		UserId:       2,
		CreatedAt:    time.Now().UTC().Add(-1 * time.Hour),
	}).Error; err != nil {
		t.Fatal(err)
	}

	if err := gormDb.Create(&database.AdCampaignView{
		AdCampaignId: 2,
		UserId:       3,
		CreatedAt:    time.Now().UTC(),
	}).Error; err != nil {
		t.Fatal(err)
	}

	if err := gormDb.Create(&database.AdCampaignClick{
		AdCampaignId: 2,
		UserId:       2,
		CreatedAt:    time.Now().UTC(),
	}).Error; err != nil {
		t.Fatal(err)
	}

	resp, err := adCampaignService.ListAdCampaigns(1, ListAdCampaignsRequest{
		Name:   null.StringFrom("ad"),
		Status: nil,
		Limit:  10,
		Offset: 0,
	}, gormDb, context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	a := assert.New(t)

	a.True(resp.TotalCount.Valid)
	a.Equal(int64(2), resp.TotalCount.Int64)
	a.Len(resp.Items, 2)
	a.Equal(int64(1), resp.Items[0].AdCampaignId)
	a.Equal(database.AdCampaignStatusModerated, resp.Items[0].Status)
	a.Equal(0, resp.Items[0].Views)
	a.Equal(0, resp.Items[0].Clicks)
	a.Equal(int64(2), resp.Items[1].AdCampaignId)
	a.Equal(database.AdCampaignStatusActive, resp.Items[1].Status)
	a.Equal(2, resp.Items[1].Views)
	a.Equal(1, resp.Items[1].Clicks)

	status := database.AdCampaignStatusActive
	resp, err = adCampaignService.ListAdCampaigns(1, ListAdCampaignsRequest{
		DateFrom: null.TimeFrom(time.Now().UTC().Add(-70 * time.Minute)),
		DateTo:   null.TimeFrom(time.Now().UTC().Add(-30 * time.Minute)),
		Status:   &status,
		Limit:    10,
		Offset:   0,
	}, gormDb, context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	a.True(resp.TotalCount.Valid)
	a.Equal(int64(1), resp.TotalCount.Int64)
	a.Len(resp.Items, 1)
	a.Equal(int64(2), resp.Items[0].AdCampaignId)
	a.Equal(database.AdCampaignStatusActive, resp.Items[0].Status)
	a.Equal(1, resp.Items[0].Views)
	a.Equal(0, resp.Items[0].Clicks)
}

func TestService_HasAdCampaigns(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresAllTables(configs.GetConfig().MasterDb, nil, nil); err != nil {
		t.Fatal(err)
	}

	resp, err := adCampaignService.HasAdCampaigns(1, gormDb)
	if err != nil {
		t.Fatal(err)
	}

	a := assert.New(t)

	a.NotNil(resp)
	a.False(resp.HasAdCampaign)

	if err = gormDb.Create(&database.AdCampaign{
		Id:        1,
		UserId:    1,
		Name:      "ad1",
		AdType:    database.AdTypeContent,
		Status:    database.AdCampaignStatusModerated,
		ContentId: 1,
		Country:   null.StringFrom("us"),
		Budget:    decimal.NewFromInt(30),
		Price:     decimal.NewFromInt(10),
		CreatedAt: time.Now().UTC().Add(-1 * time.Hour),
	}).Error; err != nil {
		t.Fatal(err)
	}

	resp, err = adCampaignService.HasAdCampaigns(1, gormDb)
	if err != nil {
		t.Fatal(err)
	}

	a.NotNil(resp)
	a.True(resp.HasAdCampaign)
}
