package template

import (
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/digitalmonsters/notification-handler/configs"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"os"
	"strings"
	"testing"
	"time"
)

var gormDb *gorm.DB
var config configs.Settings
var templateService *service

func TestMain(m *testing.M) {
	config = configs.GetConfig()
	gormDb = database.GetDb(database.DbTypeMaster)
	templateService = NewService().(*service)

	os.Exit(m.Run())
}

func TestService_EditTemplate(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresAllTables(config.MasterDb, nil, t); err != nil {
		t.Fatal(err)
	}

	template := database.RenderTemplate{Id: "1"}
	if err := gormDb.Create(&template).Error; err != nil {
		t.Fatal(err)
	}

	if err := templateService.EditTemplate(EditTemplateRequest{
		Id:       template.Id,
		Title:    "1",
		Body:     "1",
		Headline: "1",
		Kind:     "1",
		Route:    "1",
		ImageUrl: "1",
	}, gormDb); err != nil {
		t.Fatal(err)
	}

	if err := gormDb.Find(&template).Error; err != nil {
		t.Fatal(err)
	}

	a := assert.New(t)

	a.True(strings.Contains(template.Kind, "1"))
}

func TestService_ListTemplates(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresAllTables(config.MasterDb, nil, t); err != nil {
		t.Fatal(err)
	}

	template1 := database.RenderTemplate{
		Id:        "template1",
		Kind:      "template1",
		Route:     "template1",
		ImageUrl:  "template1",
		CreatedAt: time.Now().UTC().Add(-1 * time.Hour),
		UpdatedAt: time.Now().UTC().Add(-1 * time.Hour),
	}
	if err := gormDb.Create(&template1).Error; err != nil {
		t.Fatal(err)
	}

	template2 := database.RenderTemplate{
		Id:        "template2",
		Kind:      "template2",
		Route:     "template2",
		ImageUrl:  "template2",
		CreatedAt: time.Now().UTC().Add(-1 * time.Hour),
		UpdatedAt: time.Now().UTC().Add(-1 * time.Hour),
	}
	if err := gormDb.Create(&template2).Error; err != nil {
		t.Fatal(err)
	}

	resp, err := templateService.ListTemplates(ListTemplatesRequest{
		Id: "template1",
		//Title:    "template1",
		//Body:     "template1",
		//Headline: "template1",
		Kind:     "template1",
		Route:    "template1",
		ImageUrl: "template1",
		Sorting: []Sorting{
			{
				Field:       "id",
				IsAscending: true,
			},
		},
		CreatedAtFrom: null.TimeFrom(time.Now().UTC().Add(-2 * time.Hour)),
		CreatedAtTo:   null.TimeFrom(time.Now().UTC().Add(2 * time.Hour)),
		UpdatedAtFrom: null.TimeFrom(time.Now().UTC().Add(-2 * time.Hour)),
		UpdatedAtTo:   null.TimeFrom(time.Now().UTC().Add(2 * time.Hour)),
		Limit:         10,
		Offset:        0,
	}, gormDb)
	if err != nil {
		t.Fatal(err)
	}

	a := assert.New(t)

	a.NotNil(resp)
	a.Equal(int64(1), resp.TotalCount.ValueOrZero())
	a.Len(resp.Items, 1)
	a.Equal(template1.Id, resp.Items[0].Id)
}
