package settings

import (
	"context"
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/digitalmonsters/notification-handler/configs"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/gocql/gocql"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"os"
	"strings"
	"testing"
)

var gormDb *gorm.DB
var config configs.Settings
var session *gocql.Session
var settingsService *service

func TestMain(m *testing.M) {
	config = configs.GetConfig()
	session = database.GetScyllaSession()
	settingsService = NewService().(*service)
	gormDb = database.GetDb(database.DbTypeMaster)

	os.Exit(m.Run())
}

func TestService_GetPushSettings(t *testing.T) {
	if err := boilerplate_testing.FlushScyllaAllTables(nil, session, config.Scylla.Keyspace, nil); err != nil {
		t.Fatal(err)
	}
	if err := boilerplate_testing.FlushPostgresAllTables(config.MasterDb, nil, t); err != nil {
		t.Fatal(err)
	}

	settingsService.templatesCache.Flush()

	resp, err := settingsService.GetPushSettings(1, context.TODO(), gormDb)
	if err != nil {
		t.Fatal(err)
	}

	a := assert.New(t)

	a.Equal(0, len(resp))

	renderTemplate1 := database.RenderTemplate{Id: "1"}
	if err = gormDb.Create(&renderTemplate1).Error; err != nil {
		t.Fatal(err)
	}

	settingsService.templatesCache.Flush()

	resp, err = settingsService.GetPushSettings(1, context.TODO(), gormDb)
	if err != nil {
		t.Fatal(err)
	}

	a.Equal(1, len(resp))

	var ok bool
	var respVal bool

	respVal, ok = resp["1"]
	a.True(ok)
	a.False(respVal)

	if err = session.Query("update user_notifications_settings set muted = true " +
		"where cluster_key = 0 and user_id = 1 and template_id = '1'").Exec(); err != nil {
		t.Fatal(err)
	}

	resp, err = settingsService.GetPushSettings(1, context.TODO(), gormDb)
	if err != nil {
		t.Fatal(err)
	}

	a.Equal(1, len(resp))

	respVal, ok = resp["1"]
	a.True(ok)
	a.True(respVal)
}

func TestService_GetPushSettingsByAdmin(t *testing.T) {
	if err := boilerplate_testing.FlushScyllaAllTables(nil, session, config.Scylla.Keyspace, nil); err != nil {
		t.Fatal(err)
	}
	if err := boilerplate_testing.FlushPostgresAllTables(config.MasterDb, nil, t); err != nil {
		t.Fatal(err)
	}

	settingsService.templatesCache.Flush()

	resp, err := settingsService.GetPushSettingsByAdmin(GetPushSettingsByAdminRequest{UserId: 1}, context.TODO(), gormDb)
	if err != nil {
		t.Fatal(err)
	}

	a := assert.New(t)

	a.Equal(0, len(resp))

	renderTemplate1 := database.RenderTemplate{Id: "1"}
	if err = gormDb.Create(&renderTemplate1).Error; err != nil {
		t.Fatal(err)
	}

	settingsService.templatesCache.Flush()

	resp, err = settingsService.GetPushSettingsByAdmin(GetPushSettingsByAdminRequest{UserId: 1}, context.TODO(), gormDb)
	if err != nil {
		t.Fatal(err)
	}

	a.Equal(1, len(resp))

	var ok bool
	var respVal GetPushSettingsByAdminItem

	respVal, ok = resp["1"]
	a.True(ok)
	a.True(strings.Contains("1", respVal.Id))
	a.False(respVal.Muted)

	if err = session.Query("update user_notifications_settings set muted = true " +
		"where cluster_key = 0 and user_id = 1 and template_id = '1'").Exec(); err != nil {
		t.Fatal(err)
	}

	resp, err = settingsService.GetPushSettingsByAdmin(GetPushSettingsByAdminRequest{UserId: 1}, context.TODO(), gormDb)
	if err != nil {
		t.Fatal(err)
	}

	a.Equal(1, len(resp))

	respVal, ok = resp["1"]
	a.True(ok)
	a.True(strings.Contains("1", respVal.Id))
	a.True(respVal.Muted)
}

func TestService_ChangePushSettings(t *testing.T) {
	if err := boilerplate_testing.FlushScyllaAllTables(nil, session, config.Scylla.Keyspace, nil); err != nil {
		t.Fatal(err)
	}

	if err := settingsService.ChangePushSettings(map[string]bool{
		"1": true,
		"2": false,
	}, 1, context.TODO()); err != nil {
		t.Fatal(err)
	}

	iter := session.Query("select template_id, muted from user_notifications_settings where "+
		"cluster_key = ? and user_id = ?", database.GetUserNotificationsSettingsClusterKey(1), 1).Iter()

	settingsMap := make(map[string]bool)

	var templateId string
	var muted bool

	for iter.Scan(&templateId, &muted) {
		settingsMap[templateId] = muted
	}

	if err := iter.Close(); err != nil {
		t.Fatal(err)
	}

	a := assert.New(t)

	a.Equal(2, len(settingsMap))

	var ok bool
	var respVal bool

	respVal, ok = settingsMap["1"]
	a.True(ok)
	a.True(respVal)

	respVal, ok = settingsMap["2"]
	a.True(ok)
	a.False(respVal)
}

func TestService_IsPushNotificationMuted(t *testing.T) {
	if err := boilerplate_testing.FlushScyllaAllTables(nil, session, config.Scylla.Keyspace, nil); err != nil {
		t.Fatal(err)
	}

	isMuted, err := settingsService.IsPushNotificationMuted(1, "1", context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	a := assert.New(t)

	a.False(isMuted)

	if err = session.Query("update user_notifications_settings set muted = true " +
		"where cluster_key = 0 and user_id = 1 and template_id = '1'").Exec(); err != nil {
		t.Fatal(err)
	}

	isMuted, err = settingsService.IsPushNotificationMuted(1, "1", context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	a.True(isMuted)
}
