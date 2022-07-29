package common

import (
	"github.com/digitalmonsters/ads-manager/configs"
	"github.com/digitalmonsters/ads-manager/pkg/database"
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"os"
	"testing"
)

var gormDb *gorm.DB
var commonService IService
var cfg configs.Settings

func TestMain(m *testing.M) {
	cfg = configs.GetConfig()
	gormDb = database.GetDb(database.DbTypeMaster)
	commonService = NewService()

	os.Exit(m.Run())
}

func TestNewService_RejectReasons(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresAllTables(cfg.MasterDb, nil, nil); err != nil {
		t.Fatal(err)
	}
	var req = UpsertRejectReasonsRequest{Items: []UpsertRejectReason{
		{
			Id:     null.Int{},
			Reason: "test reason 1",
		},
		{
			Id:     null.Int{},
			Reason: "test reason 2",
		},
		{
			Id:     null.Int{},
			Reason: "test reason 3",
		},
	}}

	err := commonService.UpsertRejectReasons(req, gormDb)
	if err != nil {
		t.Fatal(err)
	}
	var reasons []database.RejectReason
	if err := gormDb.Order("id").Find(&reasons).Error; err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 3, len(reasons))

	resp, err := commonService.ListRejectReasons(ListRejectReasonsRequest{
		Limit:  10,
		Offset: 0,
	}, gormDb)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int64(3), resp.TotalCount)
	assert.Equal(t, 3, len(resp.Items))

	req = UpsertRejectReasonsRequest{Items: []UpsertRejectReason{
		{
			Id:     null.IntFrom(reasons[0].Id),
			Reason: "updated test reason 1",
		},
	}}

	err = commonService.UpsertRejectReasons(req, gormDb)
	if err != nil {
		t.Fatal(err)
	}
	if err := gormDb.Order("id").Find(&reasons).Error; err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 3, len(reasons))
	assert.Equal(t, "updated test reason 1", reasons[0].Reason)

	err = commonService.DeleteRejectReasons(DeleteRequest{Ids: []int64{
		reasons[1].Id, reasons[2].Id,
	}}, gormDb)
	if err != nil {
		t.Fatal(err)
	}
	if err := gormDb.Order("id").Find(&reasons).Error; err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(reasons))
}

func TestNewService_ActionButtons(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresAllTables(cfg.MasterDb, nil, nil); err != nil {
		t.Fatal(err)
	}
	var req = UpsertActionButtonsRequest{Items: []UpsertButtonItem{
		{
			Id:   null.Int{},
			Name: "test button 1",
			Type: database.ContentButtonType,
		},
		{
			Id:   null.Int{},
			Name: "test button 2",
			Type: database.LinkButtonType,
		},
		{
			Id:   null.Int{},
			Name: "test button 3",
			Type: database.ContentButtonType,
		},
	}}

	err := commonService.UpsertActionButtons(req, gormDb)
	if err != nil {
		t.Fatal(err)
	}
	var buttons []database.ActionButton
	if err := gormDb.Order("id").Find(&buttons).Error; err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 3, len(buttons))

	resp, err := commonService.ListActionButtons(ListActionButtonsRequest{
		Limit:  10,
		Offset: 0,
	}, gormDb)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int64(3), resp.TotalCount)
	assert.Equal(t, 3, len(resp.Items))

	req = UpsertActionButtonsRequest{Items: []UpsertButtonItem{
		{
			Id:   null.IntFrom(buttons[0].Id),
			Name: "updated test name 1",
			Type: database.LinkButtonType,
		},
	}}

	err = commonService.UpsertActionButtons(req, gormDb)
	if err != nil {
		t.Fatal(err)
	}
	if err := gormDb.Order("id").Find(&buttons).Error; err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 3, len(buttons))
	assert.Equal(t, "updated test name 1", buttons[0].Name)
	assert.Equal(t, database.LinkButtonType, buttons[0].Type)

	err = commonService.DeleteActionButtons(DeleteRequest{Ids: []int64{
		buttons[1].Id, buttons[2].Id,
	}}, gormDb)
	if err != nil {
		t.Fatal(err)
	}
	if err := gormDb.Order("id").Find(&buttons).Error; err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(buttons))
}
