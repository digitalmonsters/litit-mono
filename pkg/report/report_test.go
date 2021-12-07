package report

import (
	"github.com/digitalmonsters/comments/configs"
	"github.com/digitalmonsters/comments/pkg/database"
	"github.com/digitalmonsters/comments/utils"
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"os"
	"testing"
)

var db *gorm.DB

func TestMain(m *testing.M) {
	db = database.GetDb()
	os.Exit(m.Run())
}

func baseSetup(t *testing.T) {
	cfg := configs.GetConfig()

	if err := boilerplate_testing.FlushPostgresTables(cfg.Db.ToBoilerplate(),
		[]string{"public.comment", "public.comment_vote", "public.content", "public.report", "public.profile"}, nil, nil); err != nil {
		t.Fatal(err)
	}

	if err := utils.PollutePostgresDatabase(db, "../comments/test_data/seed.json"); err != nil {
		t.Fatal(err)
	}
}

func TestReportComment(t *testing.T) {
	baseSetup(t)
	report, err := ReportComment(9700, "spam", db, 1,"type")
	if err != nil {
		t.Fatal(err)
	}
	var dbReport *database.Report
	if err := db.Where("id = ?", report.Id).First(&dbReport).Error; err != nil  {
		t.Fatal(err)
	}

	a := assert.New(t)
	a.Equal(int64(9700),report.CommentId)
	a.Equal(int64(1), report.ReporterId)
	a.Equal(int64(1017738), report.ContentId)
	a.Equal("comment", report.ReportType)
	a.Equal("type", report.Type)
	a.Equal("spam", report.Detail)

	secondReport, err := ReportComment(9700, "spam", db, 1, "type")
	if err != nil {
		t.Fatal(err)
	}

	a.Equal(report.Id, secondReport.Id)
}
