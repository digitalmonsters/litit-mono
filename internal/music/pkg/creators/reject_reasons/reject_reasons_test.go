package reject_reasons

import (
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/digitalmonsters/music/configs"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"os"
	"testing"
)

var config configs.Settings
var gormDb *gorm.DB

func TestMain(m *testing.M) {
	config = configs.GetConfig()
	gormDb = database.GetDb(database.DbTypeMaster)
	os.Exit(m.Run())
}

func TestMethods(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresTables(config.MasterDb, []string{"public.creator_reject_reasons"}, nil, t); err != nil {
		t.Fatal(err)
	}

	upsertReq := []rejectReason{
		{
			Reason: "test_reason_1",
		},
		{
			Reason: "test_reason_2",
		},
	}

	resp, err := Upsert(UpsertRequest{
		Items: upsertReq,
	}, gormDb)

	assert.Nil(t, err)
	assert.Len(t, resp, 2)

	list, err := List(ListRequest{
		Limit: 10,
	}, gormDb)

	assert.Nil(t, err)
	assert.Len(t, list.Items, 2)

	err = Delete(DeleteRequest{Ids: []int64{resp[0].Id}}, gormDb)
	assert.Nil(t, err)

	list, err = List(ListRequest{
		Limit: 10,
	}, gormDb)
	assert.Nil(t, err)
	assert.Len(t, list.Items, 1)
	assert.Equal(t, list.Items[0].Reason, upsertReq[1].Reason)
}
