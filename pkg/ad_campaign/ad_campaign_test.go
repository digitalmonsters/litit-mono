package ad_campaign

import (
	"context"
	"github.com/digitalmonsters/ads-manager/configs"
	"github.com/digitalmonsters/ads-manager/pkg/database"
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/digitalmonsters/go-common/wrappers"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/stretchr/testify/assert"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"os"
	"strings"
	"testing"
)

var gormDb *gorm.DB
var adCampaignService IService
var contentWrapperMock content.IContentWrapper

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
	}

	adCampaignService = NewService(contentWrapperMock)

	os.Exit(m.Run())
}

func TestService_CreateAdCampaign(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresAllTables(configs.GetConfig().MasterDb, nil, nil); err != nil {
		t.Fatal(err)
	}

	if err := adCampaignService.CreateAdCampaign(CreateAdCampaignRequest{
		Name:         "a1",
		AdType:       database.AdTypeContent,
		ContentId:    1,
		Link:         null.StringFrom("l1"),
		LinkButtonId: null.IntFrom(1),
		Country:      null.StringFrom("usa"),
		DurationMin:  1,
		Budget:       1,
		Gender:       null.StringFrom("male"),
		AgeFrom:      0,
		AgeTo:        100,
	}, 1, gormDb, context.TODO()); err != nil {
		t.Fatal(err)
	}

	var adCampaign database.AdCampaign
	if err := gormDb.First(&adCampaign).Error; err != nil {
		t.Fatal(err)
	}

	a := assert.New(t)

	a.True(strings.EqualFold("a1", adCampaign.Name))
}
