package template

import (
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/digitalmonsters/notification-handler/configs"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"os"
	"strings"
	"testing"
)

var gormDb *gorm.DB
var config configs.Settings
var templateService *service

func TestMain(m *testing.M) {
	config = configs.GetConfig()
	templateService = NewService().(*service)
	gormDb = database.GetDb(database.DbTypeMaster)

	os.Exit(m.Run())
}

func TestService_EditTemplate(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresAllTables(config.MasterDb, nil, t); err != nil {
		t.Fatal(err)
	}

	template := database.RenderTemplate{Id: "1"}
	if err := gormDb.Create(&template); err != nil {
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

	a.True(strings.Contains(template.Title, "1"))
}

func TestService_ListTemplates(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresAllTables(config.MasterDb, nil, t); err != nil {
		t.Fatal(err)
	}
}
