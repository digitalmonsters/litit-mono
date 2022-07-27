package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/digitalmonsters/go-common/wrappers"
	"github.com/digitalmonsters/go-common/wrappers/follow"
	"github.com/digitalmonsters/go-common/wrappers/notification_handler"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/notification-handler/configs"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/database/scylla"
	"github.com/gocql/gocql"
	"github.com/stretchr/testify/assert"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"os"
	"testing"
	"time"
)

var config configs.Settings
var gormDb *gorm.DB
var session *gocql.Session
var userWrapperMock *user_go.UserGoWrapperMock
var followWrapper *follow.FollowWrapperMock

func TestMain(m *testing.M) {
	config = configs.GetConfig()
	session = database.GetScyllaSession()
	gormDb = database.GetDb(database.DbTypeMaster)

	userWrapperMock = &user_go.UserGoWrapperMock{
		GetUsersDetailFn: func(userIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]user_go.UserDetailRecord] {
			respChan := make(chan wrappers.GenericResponseChan[map[int64]user_go.UserDetailRecord], 2)
			go func() {
				respMap := make(map[int64]user_go.UserDetailRecord)

				for i, userId := range userIds {
					respMap[userId] = user_go.UserDetailRecord{
						Id:                userId,
						Username:          null.StringFrom(fmt.Sprintf("%vusername", i)),
						Firstname:         fmt.Sprintf("%vfirstname", i),
						Lastname:          fmt.Sprintf("%vlastname", i),
						Followers:         i,
						Avatar:            null.StringFrom(fmt.Sprintf("%vusername", i)),
						NamePrivacyStatus: user_go.NamePrivacyStatusVisible,
					}
				}

				respChan <- wrappers.GenericResponseChan[map[int64]user_go.UserDetailRecord]{
					Response: respMap,
				}
			}()
			return respChan
		},
		GetUserBlockFn: func(blockedTo int64, blockedBy int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[user_go.UserBlockData] {
			respChan := make(chan wrappers.GenericResponseChan[user_go.UserBlockData], 2)
			respChan <- wrappers.GenericResponseChan[user_go.UserBlockData]{
				Response: user_go.UserBlockData{
					Type:      nil,
					IsBlocked: false,
				},
			}

			return respChan
		},
	}

	followWrapper = &follow.FollowWrapperMock{}
	followWrapper.GetUserFollowingRelationBulkFn = func(userId int64, requestUserIds []int64, apmTransaction *apm.Transaction,
		forceLog bool) chan follow.GetUserFollowingRelationBulkResponseChan {
		ch := make(chan follow.GetUserFollowingRelationBulkResponseChan, 2)

		ch <- follow.GetUserFollowingRelationBulkResponseChan{
			Error: nil,
			Data:  map[int64]follow.RelationData{},
		}
		close(ch)

		return ch
	}

	os.Exit(m.Run())
}

func TestService_GetNotifications(t *testing.T) {
	if err := boilerplate_testing.FlushScyllaAllTables(nil, session, config.Scylla.Keyspace, nil); err != nil {
		t.Fatal(err)
	}
	if err := boilerplate_testing.FlushPostgresAllTables(config.MasterDb, nil, t); err != nil {
		t.Fatal(err)
	}

	dbNotification1 := database.Notification{
		UserId:        1,
		Type:          "push.profile.following",
		Title:         "1",
		Message:       "1",
		RelatedUserId: null.IntFrom(1),
		CreatedAt:     time.Now().UTC().Add(-10 * time.Minute),
	}
	if err := gormDb.Create(&dbNotification1).Error; err != nil {
		t.Fatal(err)
	}

	dbNotification1Marshalled, err := json.Marshal(dbNotification1)
	if err != nil {
		t.Fatal(err)
	}

	notification1 := scylla.Notification{
		UserId:             dbNotification1.UserId,
		EventType:          "follow",
		EntityId:           dbNotification1.RelatedUserId.ValueOrZero(),
		CreatedAt:          dbNotification1.CreatedAt,
		NotificationsCount: 1,
		Title:              dbNotification1.Title,
		Body:               dbNotification1.Message,
		Headline:           dbNotification1.Title,
		Kind:               "1",
		NotificationInfo:   string(dbNotification1Marshalled),
	}

	dbNotification2 := database.Notification{
		UserId:        1,
		Type:          "push.profile.following",
		Title:         "2",
		Message:       "2",
		RelatedUserId: null.IntFrom(2),
		CreatedAt:     time.Now().UTC(),
	}
	if err = gormDb.Create(&dbNotification2).Error; err != nil {
		t.Fatal(err)
	}

	dbNotification2Marshalled, err := json.Marshal(dbNotification2)
	if err != nil {
		t.Fatal(err)
	}

	notification2 := scylla.Notification{
		UserId:             dbNotification2.UserId,
		EventType:          "follow",
		EntityId:           dbNotification2.RelatedUserId.ValueOrZero(),
		CreatedAt:          dbNotification2.CreatedAt,
		NotificationsCount: 2,
		Title:              dbNotification2.Title,
		Body:               dbNotification2.Message,
		Headline:           dbNotification2.Title,
		Kind:               "1",
		NotificationInfo:   string(dbNotification2Marshalled),
	}

	if err = session.Query("update notification set notifications_count = ?, title = ?, body = ?, headline = ?, kind = ?, rendering_variables = ?, "+
		"custom_data = ?, notification_info = ? where user_id = ? and event_type = ? "+
		"and created_at = ? and entity_id = ? and related_entity_id = ?", notification1.NotificationsCount, notification1.Title, notification1.Body, notification1.Headline,
		notification1.Kind, notification1.RenderingVariables, notification1.CustomData, notification1.NotificationInfo,
		notification1.UserId, notification1.EventType, notification1.CreatedAt, notification1.EntityId, notification1.RelatedEntityId).Exec(); err != nil {
		t.Fatal(err)
	}

	if err = session.Query("update notification set notifications_count = ?, title = ?, body = ?, headline = ?, kind = ?, rendering_variables = ?, "+
		"custom_data = ?, notification_info = ? where user_id = ? and event_type = ? "+
		"and created_at = ? and entity_id = ? and related_entity_id = ?", notification2.NotificationsCount, notification2.Title, notification2.Body, notification2.Headline,
		notification2.Kind, notification2.RenderingVariables, notification2.CustomData, notification2.NotificationInfo,
		notification2.UserId, notification2.EventType, notification2.CreatedAt, notification2.EntityId, notification2.RelatedEntityId).Exec(); err != nil {
		t.Fatal(err)
	}

	resp, err := GetNotifications(gormDb, dbNotification2.UserId, "", TypeGroupAll, true, 1, userWrapperMock, followWrapper, context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	a := assert.New(t)

	a.NotNil(resp)
	a.Len(resp.Data, 1)
	a.Equal(dbNotification2.Id, resp.Data[0].Id)

	resp, err = GetNotifications(gormDb, dbNotification2.UserId, resp.Next, TypeGroupAll, true, 1, userWrapperMock, followWrapper, context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	a.NotNil(resp)
	a.Len(resp.Data, 1)
	a.Equal(dbNotification1.Id, resp.Data[0].Id)

	resp, err = GetNotifications(gormDb, dbNotification2.UserId, resp.Next, TypeGroupAll, false, 10, userWrapperMock, followWrapper, context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	a.NotNil(resp)
	a.Len(resp.Data, 0)

	dbNotification3 := database.Notification{
		UserId:    1,
		Type:      "push.admin.bulk",
		Title:     "4",
		Message:   "4",
		CreatedAt: time.Now().UTC(),
	}
	if err = gormDb.Create(&dbNotification3).Error; err != nil {
		t.Fatal(err)
	}

	dbNotification3Marshalled, err := json.Marshal(dbNotification3)
	if err != nil {
		t.Fatal(err)
	}

	notification3 := scylla.Notification{
		UserId:             dbNotification3.UserId,
		EventType:          "push_admin",
		CreatedAt:          dbNotification2.CreatedAt,
		NotificationsCount: 1,
		Title:              dbNotification3.Title,
		Body:               dbNotification3.Message,
		Headline:           dbNotification3.Title,
		Kind:               "1",
		NotificationInfo:   string(dbNotification3Marshalled),
	}
	if err = session.Query("update notification set notifications_count = ?, title = ?, body = ?, headline = ?, kind = ?, rendering_variables = ?, "+
		"custom_data = ?, notification_info = ? where user_id = ? and event_type = ? "+
		"and created_at = ? and entity_id = ? and related_entity_id = ?", notification3.NotificationsCount, notification3.Title, notification3.Body, notification3.Headline,
		notification3.Kind, notification3.RenderingVariables, notification3.CustomData, notification3.NotificationInfo,
		notification3.UserId, notification3.EventType, notification3.CreatedAt, notification3.EntityId, notification3.RelatedEntityId).Exec(); err != nil {
		t.Fatal(err)
	}

	resp, err = GetNotifications(gormDb, dbNotification3.UserId, "", TypeGroupSystem, true, 1, userWrapperMock, followWrapper, context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	a.NotNil(resp)
	a.Len(resp.Data, 1)
	a.Equal(dbNotification3.Id, resp.Data[0].Id)
}

func TestDisableUnregisteredTokens(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresAllTables(config.MasterDb, nil, t); err != nil {
		t.Fatal(err)
	}
	tokens := make([]string, 0)
	for i := int64(1); i <= 10; i++ {
		token := "push_token" + fmt.Sprint(i)
		tokens = append(tokens, token)
		if err := gormDb.Create(&database.Device{
			UserId:       i,
			DeviceId:     "device_id" + fmt.Sprint(i),
			PushToken:    token,
			Platform:     "some_platform",
			Unregistered: false,
		}).Error; err != nil {
			t.Fatal(err)
		}
	}

	if _, err := DisableUnregisteredTokens(notification_handler.DisableUnregisteredTokensRequest{
		Tokens: tokens[:5],
	}, gormDb); err != nil {
		t.Fatal(err)
	}

	var devices []database.Device

	if err := gormDb.Where("unregistered = true").Find(&devices).Error; err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 5, len(devices))

}
