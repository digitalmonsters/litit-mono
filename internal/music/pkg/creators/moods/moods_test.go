package moods

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
	if err := boilerplate_testing.FlushPostgresTables(config.MasterDb, []string{"public.creator_reject_reasons", "public.moods"}, nil, t); err != nil {
		t.Fatal(err)
	}

	upsertReq := []mood{
		{
			Name:      "test_mood_1",
			SortOrder: 1,
		},
		{
			Name:      "test_mood_2",
			SortOrder: 2,
		},
	}

	resp, err := Upsert(UpsertRequest{
		Items: upsertReq,
	}, gormDb)

	assert.Nil(t, err)
	assert.Len(t, resp, 2)

	list, err := AdminList(ListRequest{
		Limit: 10,
	}, gormDb)

	assert.Nil(t, err)
	assert.Len(t, list.Items, 2)

	err = Delete(DeleteRequest{Ids: []int64{resp[0].Id}}, gormDb)
	assert.Nil(t, err)

	list, err = AdminList(ListRequest{
		Limit: 10,
	}, gormDb)
	assert.Nil(t, err)
	assert.Len(t, list.Items, 1)
	assert.Equal(t, list.Items[0].Name, upsertReq[1].Name)
}
