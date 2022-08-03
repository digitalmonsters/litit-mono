package view_content

import (
	"context"
	"github.com/digitalmonsters/ads-manager/configs"
	"github.com/digitalmonsters/ads-manager/pkg/database"
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/wrappers"
	"github.com/digitalmonsters/go-common/wrappers/go_tokenomics"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"os"
	"testing"
)

var gormDb *gorm.DB
var config configs.Settings
var goTokenomicsWrapper *go_tokenomics.GoTokenomicsWrapperMock
var viewContentService *service

func TestMain(m *testing.M) {
	config = configs.GetConfig()
	gormDb = database.GetDb(database.DbTypeMaster)

	goTokenomicsWrapper = &go_tokenomics.GoTokenomicsWrapperMock{}

	goTokenomicsWrapper.WriteOffUserTokensForAdFn = func(userId int64, adCampaignId int64, amount decimal.Decimal, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[any] {
		var respCh = make(chan wrappers.GenericResponseChan[any], 1)
		respCh <- wrappers.GenericResponseChan[any]{
			Error:    nil,
			Response: nil,
		}
		close(respCh)
		return respCh
	}

	viewContentService = NewViewContentService(goTokenomicsWrapper).(*service)

	os.Exit(m.Run())
}

func Test_handleOne(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresAllTables(config.MasterDb, nil, t); err != nil {
		t.Fatal(err)
	}

	if err := gormDb.Create(&database.AdCampaignCountriesPrice{
		CountryCode:   "us",
		Price:         decimal.NewFromInt(10),
		CountryName:   "USA",
		IsGlobalPrice: false,
	}).Error; err != nil {
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

	if err := viewContentService.handleOne(gormDb, fullEvent{
		ViewEvent: eventsourcing.ViewEvent{
			UserId:          2,
			ContentId:       1,
			ContentType:     eventsourcing.ContentTypeVideo,
			ContentAuthorId: 1,
			SourceView:      eventsourcing.SourceViewFeedInterests,
		},
	}, context.TODO()); err != nil {
		t.Fatal(err)
	}

	var adCampaign database.AdCampaign
	if err := gormDb.Where("id = 1").Find(&adCampaign).Error; err != nil {
		t.Fatal(err)
	}

	a := assert.New(t)

	a.True(adCampaign.Budget.Equal(decimal.NewFromInt(20)))
	a.True(adCampaign.Paid)
	a.Equal(1, adCampaign.Views)

	var adCampaignView database.AdCampaignView
	if err := gormDb.Where("ad_campaign_id = 1 and user_id = 2").Find(&adCampaignView).Error; err != nil {
		t.Fatal(err)
	}

	a.Equal(int64(1), adCampaignView.AdCampaignId)
}
