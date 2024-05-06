package sender

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/translation"
	"github.com/digitalmonsters/go-common/wrappers"
	"github.com/digitalmonsters/go-common/wrappers/notification_gateway"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/notification-handler/configs"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/database/scylla"
	"github.com/digitalmonsters/notification-handler/pkg/settings"
	"github.com/gocql/gocql"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"os"
	"testing"
	"time"
)

var config configs.Settings
var gormDb *gorm.DB
var session *gocql.Session
var sender *Sender
var gateway *notification_gateway.NotificationGatewayWrapperMock
var settingsServiceMock *settings.ServiceMock
var userWrapperMock *user_go.UserGoWrapperMock
var pushSendMessages []notification_gateway.SendPushRequest

func TestMain(m *testing.M) {
	config = configs.GetConfig()
	session = database.GetScyllaSession()
	gormDb = database.GetDb(database.DbTypeMaster)
	pushSendMessages = nil

	gateway = &notification_gateway.NotificationGatewayWrapperMock{}
	gateway.EnqueuePushForUserFn = func(msg []notification_gateway.SendPushRequest, ctx context.Context) chan error {
		pushSendMessages = msg
		ch := make(chan error, 2)
		ch <- nil
		close(ch)
		return ch
	}

	settingsServiceMock = &settings.ServiceMock{}
	settingsServiceMock.IsPushNotificationMutedFn = func(userId int64, templateId string, ctx context.Context) (bool, error) {
		return false, nil
	}

	userWrapperMock = &user_go.UserGoWrapperMock{}
	userWrapperMock.GetUsersFn = func(userIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]user_go.UserRecord] {
		respChan := make(chan wrappers.GenericResponseChan[map[int64]user_go.UserRecord], 2)
		go func() {
			respMap := make(map[int64]user_go.UserRecord)

			for i, userId := range userIds {
				respMap[userId] = user_go.UserRecord{
					UserId:            userId,
					Username:          fmt.Sprintf("%vusername", i),
					Firstname:         fmt.Sprintf("%vfirstname", i),
					Lastname:          fmt.Sprintf("%vlastname", i),
					Avatar:            null.StringFrom(fmt.Sprintf("%vusername", i)),
					NamePrivacyStatus: user_go.NamePrivacyStatusVisible,
				}
			}

			respChan <- wrappers.GenericResponseChan[map[int64]user_go.UserRecord]{
				Response: respMap,
			}
		}()
		return respChan
	}

	sender = NewSender(gateway, settingsServiceMock, nil, userWrapperMock)

	os.Exit(m.Run())
}

func TestTimeToNearestMinutes(t *testing.T) {
	date := time.Date(2022, 06, 21, 17, 32, 20, 4, time.UTC)
	newDate := TimeToNearestMinutes(date, 30, true)

	a := assert.New(t)

	a.Equal(date.Year(), newDate.Year())
	a.Equal(date.Month(), newDate.Month())
	a.Equal(date.Day(), newDate.Day())
	a.Equal(date.Hour(), newDate.Hour())
	a.Equal(30, newDate.Minute())
	a.Equal(0, newDate.Second())
	a.Equal(0, newDate.Nanosecond())

	newDate = TimeToNearestMinutes(date, 30, false)

	a.Equal(date.Year(), newDate.Year())
	a.Equal(date.Month(), newDate.Month())
	a.Equal(date.Day(), newDate.Day())
	a.Equal(date.Hour()+1, newDate.Hour())
	a.Equal(0, newDate.Minute())
	a.Equal(0, newDate.Second())
	a.Equal(0, newDate.Nanosecond())

	newDate = TimeToNearestMinutes(date, 120, true)

	a.Equal(date.Year(), newDate.Year())
	a.Equal(date.Month(), newDate.Month())
	a.Equal(date.Day(), newDate.Day())
	a.Equal(date.Hour()-1, newDate.Hour())
	a.Equal(0, newDate.Minute())
	a.Equal(0, newDate.Second())
	a.Equal(0, newDate.Nanosecond())

	newDate = TimeToNearestMinutes(date, 120, false)

	a.Equal(date.Year(), newDate.Year())
	a.Equal(date.Month(), newDate.Month())
	a.Equal(date.Day(), newDate.Day())
	a.Equal(date.Hour()+1, newDate.Hour())
	a.Equal(0, newDate.Minute())
	a.Equal(0, newDate.Second())
	a.Equal(0, newDate.Nanosecond())

	newDate = TimeToNearestMinutes(date, 121, false)

	a.Equal(date.Year(), newDate.Year())
	a.Equal(date.Month(), newDate.Month())
	a.Equal(date.Day(), newDate.Day())
	a.Equal(date.Hour()+1, newDate.Hour())
	a.Equal(1, newDate.Minute())
	a.Equal(0, newDate.Second())
	a.Equal(0, newDate.Nanosecond())

	newDate = TimeToNearestMinutes(date, 21662, true)

	a.Equal(date.Year(), newDate.Year())
	a.Equal(date.Month(), newDate.Month())
	a.Equal(16, newDate.Day())
	a.Equal(1, newDate.Hour())
	a.Equal(2, newDate.Minute())
	a.Equal(0, newDate.Second())
	a.Equal(0, newDate.Nanosecond())

	newDate = TimeToNearestMinutes(date, 21662, false)

	a.Equal(date.Year(), newDate.Year())
	a.Equal(date.Month()+1, newDate.Month())
	a.Equal(1, newDate.Day())
	a.Equal(1, newDate.Hour())
	a.Equal(2, newDate.Minute())
	a.Equal(0, newDate.Second())
	a.Equal(0, newDate.Nanosecond())

	newDate = TimeToNearestMinutes(date, 302462, true)

	a.Equal(date.Year(), newDate.Year())
	a.Equal(time.Month(1), newDate.Month())
	a.Equal(1, newDate.Day())
	a.Equal(1, newDate.Hour())
	a.Equal(2, newDate.Minute())
	a.Equal(0, newDate.Second())
	a.Equal(0, newDate.Nanosecond())

	newDate = TimeToNearestMinutes(date, 302462, false)

	a.Equal(date.Year(), newDate.Year())
	a.Equal(time.Month(8), newDate.Month())
	a.Equal(1, newDate.Day())
	a.Equal(1, newDate.Hour())
	a.Equal(2, newDate.Minute())
	a.Equal(0, newDate.Second())
	a.Equal(0, newDate.Nanosecond())

	newDate = TimeToNearestMinutes(date, 561600, true)

	a.Equal(date.Year()-1, newDate.Year())
	a.Equal(time.Month(1), newDate.Month())
	a.Equal(31, newDate.Day())
	a.Equal(0, newDate.Hour())
	a.Equal(0, newDate.Minute())
	a.Equal(0, newDate.Second())
	a.Equal(0, newDate.Nanosecond())

	newDate = TimeToNearestMinutes(date, 561600, false)

	a.Equal(date.Year(), newDate.Year())
	a.Equal(time.Month(1), newDate.Month())
	a.Equal(31, newDate.Day())
	a.Equal(0, newDate.Hour())
	a.Equal(0, newDate.Minute())
	a.Equal(0, newDate.Second())
	a.Equal(0, newDate.Nanosecond())
}

func TestSender_PushNotification_Likes(t *testing.T) {
	if err := boilerplate_testing.FlushScyllaAllTables(nil, session, config.Scylla.Keyspace, nil); err != nil {
		t.Fatal(err)
	}
	if err := boilerplate_testing.FlushPostgresAllTables(config.MasterDb, nil, t); err != nil {
		t.Fatal(err)
	}

	template := database.RenderTemplate{
		Id:        "content_like",
		Kind:      "default",
		IsGrouped: true,
	}
	if err := gormDb.Create(&template).Error; err != nil {
		t.Fatal(err)
	}

	if _, err := sender.PushNotification(database.Notification{
		UserId:        1,
		Type:          "push.content.like",
		ContentId:     null.IntFrom(1),
		RelatedUserId: null.IntFrom(2),
		RenderingVariables: database.RenderingVariables{
			"firstname": "test_f2",
			"lastname":  "test_l2",
		},
	}, 1, 2, template.Id, translation.DefaultUserLanguage, "default", context.TODO()); err != nil {
		t.Fatal(err)
	}

	var dbNotification database.Notification
	if err := gormDb.First(&dbNotification).Error; err != nil {
		t.Fatal(err)
	}

	a := assert.New(t)

	a.Equal("push.content.like", dbNotification.Type)
	a.Equal(int64(1), dbNotification.UserId)
	a.Equal(int64(2), dbNotification.RelatedUserId.ValueOrZero())

	iter := session.Query("select user_id, event_type, entity_id, related_entity_id, created_at, notifications_count, " +
		"title, body, headline, kind, rendering_variables, custom_data, notification_info from notification").Iter()

	var notification scylla.Notification
	var notifications []scylla.Notification
	for iter.Scan(&notification.UserId, &notification.EventType, &notification.EntityId, &notification.RelatedEntityId,
		&notification.CreatedAt, &notification.NotificationsCount, &notification.Title, &notification.Body,
		&notification.Headline, &notification.Kind, &notification.RenderingVariables, &notification.CustomData, &notification.NotificationInfo) {
		notifications = append(notifications, notification)
	}

	if err := iter.Close(); err != nil {
		t.Fatal(err)
	}

	a.Len(notifications, 1)

	a.Equal(int64(1), notifications[0].UserId)
	a.Equal(template.Id, notifications[0].EventType)
	a.Equal(int64(1), notifications[0].EntityId)
	a.Equal(int64(2), notifications[0].RelatedEntityId)
	a.Equal(int64(1), notifications[0].NotificationsCount)

	iter = session.Query("select user_id, event_type, entity_id, related_entity_id, event_applied from notification_relation").Iter()

	var notificationRelation scylla.NotificationRelation
	var notificationRelations []scylla.NotificationRelation
	for iter.Scan(&notificationRelation.UserId, &notificationRelation.EventType, &notificationRelation.EntityId,
		&notificationRelation.RelatedEntityId, &notificationRelation.EventApplied) {
		notificationRelations = append(notificationRelations, notificationRelation)
	}

	if err := iter.Close(); err != nil {
		t.Fatal(err)
	}

	a.Len(notificationRelations, 1)
	a.Equal(int64(1), notificationRelations[0].UserId)
	a.Equal(template.Id, notificationRelations[0].EventType)
	a.Equal(int64(1), notificationRelations[0].EntityId)
	a.Equal(int64(2), notificationRelations[0].RelatedEntityId)
	a.True(notificationRelations[0].EventApplied)

	if _, err := sender.PushNotification(database.Notification{
		UserId:        1,
		Type:          "push.content.like",
		ContentId:     null.IntFrom(1),
		RelatedUserId: null.IntFrom(2),
		RenderingVariables: database.RenderingVariables{
			"firstname": "test_f3",
			"lastname":  "test_l3",
		},
	}, 1, 3, template.Id, translation.DefaultUserLanguage, "default", context.TODO()); err != nil {
		t.Fatal(err)
	}

	prevDbNotificationId := dbNotification.Id
	dbNotification = database.Notification{}
	if err := gormDb.Where("id != ?", prevDbNotificationId).First(&dbNotification).Error; err != nil {
		t.Fatal(err)
	}

	a.Equal("push.content.like", dbNotification.Type)
	a.Equal(int64(1), dbNotification.UserId)
	a.Equal(int64(2), dbNotification.RelatedUserId.ValueOrZero())

	iter = session.Query("select user_id, event_type, entity_id, related_entity_id, created_at, notifications_count, " +
		"title, body, headline, kind, rendering_variables, custom_data, notification_info from notification").Iter()

	notification = scylla.Notification{}
	notifications = []scylla.Notification{}
	for iter.Scan(&notification.UserId, &notification.EventType, &notification.EntityId, &notification.RelatedEntityId,
		&notification.CreatedAt, &notification.NotificationsCount, &notification.Title, &notification.Body,
		&notification.Headline, &notification.Kind, &notification.RenderingVariables, &notification.CustomData, &notification.NotificationInfo) {
		notifications = append(notifications, notification)
	}

	if err := iter.Close(); err != nil {
		t.Fatal(err)
	}

	a.Len(notifications, 1)

	a.Equal(int64(1), notifications[0].UserId)
	a.Equal(template.Id, notifications[0].EventType)
	a.Equal(int64(1), notifications[0].EntityId)
	a.Equal(int64(3), notifications[0].RelatedEntityId)
	a.Equal(int64(2), notifications[0].NotificationsCount)

	iter = session.Query("select user_id, event_type, entity_id, related_entity_id, event_applied from notification_relation").Iter()

	notificationRelation = scylla.NotificationRelation{}
	notificationRelations = []scylla.NotificationRelation{}
	for iter.Scan(&notificationRelation.UserId, &notificationRelation.EventType, &notificationRelation.EntityId,
		&notificationRelation.RelatedEntityId, &notificationRelation.EventApplied) {
		notificationRelations = append(notificationRelations, notificationRelation)
	}

	if err := iter.Close(); err != nil {
		t.Fatal(err)
	}

	a.Len(notificationRelations, 2)

	a.Equal(int64(1), notificationRelations[0].UserId)
	a.Equal(template.Id, notificationRelations[0].EventType)
	a.Equal(int64(1), notificationRelations[0].EntityId)
	a.Equal(int64(2), notificationRelations[0].RelatedEntityId)
	a.True(notificationRelations[0].EventApplied)

	a.Equal(int64(1), notificationRelations[1].UserId)
	a.Equal(template.Id, notificationRelations[1].EventType)
	a.Equal(int64(1), notificationRelations[1].EntityId)
	a.Equal(int64(3), notificationRelations[1].RelatedEntityId)
	a.True(notificationRelations[1].EventApplied)
}

func TestSender_PushNotification_Follows(t *testing.T) {
	if err := boilerplate_testing.FlushScyllaAllTables(nil, session, config.Scylla.Keyspace, nil); err != nil {
		t.Fatal(err)
	}
	if err := boilerplate_testing.FlushPostgresAllTables(config.MasterDb, nil, t); err != nil {
		t.Fatal(err)
	}

	template := database.RenderTemplate{
		Id:        "follow",
		Kind:      "default",
		IsGrouped: true,
	}
	if err := gormDb.Create(&template).Error; err != nil {
		t.Fatal(err)
	}

	if _, err := sender.PushNotification(database.Notification{
		UserId:        1,
		Type:          "push.profile.following",
		RelatedUserId: null.IntFrom(2),
		RenderingVariables: database.RenderingVariables{
			"firstname": "test_f2",
			"lastname":  "test_l2",
		},
	}, 2, 0, template.Id, translation.DefaultUserLanguage, "user_follow", context.TODO()); err != nil {
		t.Fatal(err)
	}

	var dbNotification database.Notification
	if err := gormDb.First(&dbNotification).Error; err != nil {
		t.Fatal(err)
	}

	a := assert.New(t)

	a.Equal("push.profile.following", dbNotification.Type)
	a.Equal(int64(1), dbNotification.UserId)
	a.Equal(int64(2), dbNotification.RelatedUserId.ValueOrZero())

	iter := session.Query("select user_id, event_type, entity_id, related_entity_id, created_at, notifications_count, " +
		"title, body, headline, kind, rendering_variables, custom_data, notification_info from notification").Iter()

	var notification scylla.Notification
	var notifications []scylla.Notification
	for iter.Scan(&notification.UserId, &notification.EventType, &notification.EntityId, &notification.RelatedEntityId,
		&notification.CreatedAt, &notification.NotificationsCount, &notification.Title, &notification.Body,
		&notification.Headline, &notification.Kind, &notification.RenderingVariables, &notification.CustomData, &notification.NotificationInfo) {
		notifications = append(notifications, notification)
	}

	if err := iter.Close(); err != nil {
		t.Fatal(err)
	}

	a.Len(notifications, 1)

	a.Equal(int64(1), notifications[0].UserId)
	a.Equal(template.Id, notifications[0].EventType)
	a.Equal(int64(2), notifications[0].EntityId)
	a.Equal(int64(0), notifications[0].RelatedEntityId)
	a.Equal(int64(1), notifications[0].NotificationsCount)

	iter = session.Query("select user_id, event_type, entity_id, related_entity_id, event_applied from notification_relation").Iter()

	var notificationRelation scylla.NotificationRelation
	var notificationRelations []scylla.NotificationRelation
	for iter.Scan(&notificationRelation.UserId, &notificationRelation.EventType, &notificationRelation.EntityId,
		&notificationRelation.RelatedEntityId, &notificationRelation.EventApplied) {
		notificationRelations = append(notificationRelations, notificationRelation)
	}

	if err := iter.Close(); err != nil {
		t.Fatal(err)
	}

	a.Len(notificationRelations, 1)
	a.Equal(int64(1), notificationRelations[0].UserId)
	a.Equal(template.Id, notificationRelations[0].EventType)
	a.Equal(int64(2), notificationRelations[0].EntityId)
	a.Equal(int64(0), notificationRelations[0].RelatedEntityId)
	a.True(notificationRelations[0].EventApplied)

	if _, err := sender.PushNotification(database.Notification{
		UserId:        1,
		Type:          "push.profile.following",
		RelatedUserId: null.IntFrom(3),
		RenderingVariables: database.RenderingVariables{
			"firstname": "test_f3",
			"lastname":  "test_l3",
		},
	}, 3, 0, template.Id, translation.DefaultUserLanguage, "user_follow", context.TODO()); err != nil {
		t.Fatal(err)
	}

	prevDbNotificationId := dbNotification.Id
	dbNotification = database.Notification{}
	if err := gormDb.Where("id != ?", prevDbNotificationId).First(&dbNotification).Error; err != nil {
		t.Fatal(err)
	}

	a.Equal("push.profile.following", dbNotification.Type)
	a.Equal(int64(1), dbNotification.UserId)
	a.Equal(int64(3), dbNotification.RelatedUserId.ValueOrZero())

	iter = session.Query("select user_id, event_type, entity_id, related_entity_id, created_at, notifications_count, " +
		"title, body, headline, kind, rendering_variables, custom_data, notification_info from notification").Iter()

	notification = scylla.Notification{}
	notifications = []scylla.Notification{}
	for iter.Scan(&notification.UserId, &notification.EventType, &notification.EntityId, &notification.RelatedEntityId,
		&notification.CreatedAt, &notification.NotificationsCount, &notification.Title, &notification.Body,
		&notification.Headline, &notification.Kind, &notification.RenderingVariables, &notification.CustomData, &notification.NotificationInfo) {
		notifications = append(notifications, notification)
	}

	if err := iter.Close(); err != nil {
		t.Fatal(err)
	}

	a.Len(notifications, 1)

	a.Equal(int64(1), notifications[0].UserId)
	a.Equal(template.Id, notifications[0].EventType)
	a.Equal(int64(3), notifications[0].EntityId)
	a.Equal(int64(0), notifications[0].RelatedEntityId)
	a.Equal(int64(2), notifications[0].NotificationsCount)

	iter = session.Query("select user_id, event_type, entity_id, related_entity_id, event_applied from notification_relation").Iter()

	notificationRelation = scylla.NotificationRelation{}
	notificationRelations = []scylla.NotificationRelation{}
	for iter.Scan(&notificationRelation.UserId, &notificationRelation.EventType, &notificationRelation.EntityId,
		&notificationRelation.RelatedEntityId, &notificationRelation.EventApplied) {
		notificationRelations = append(notificationRelations, notificationRelation)
	}

	if err := iter.Close(); err != nil {
		t.Fatal(err)
	}

	a.Len(notificationRelations, 2)

	a.Equal(int64(1), notificationRelations[0].UserId)
	a.Equal(template.Id, notificationRelations[0].EventType)
	a.Equal(int64(2), notificationRelations[0].EntityId)
	a.Equal(int64(0), notificationRelations[0].RelatedEntityId)
	a.True(notificationRelations[0].EventApplied)

	a.Equal(int64(1), notificationRelations[1].UserId)
	a.Equal(template.Id, notificationRelations[1].EventType)
	a.Equal(int64(3), notificationRelations[1].EntityId)
	a.Equal(int64(0), notificationRelations[1].RelatedEntityId)
	a.True(notificationRelations[1].EventApplied)
}

func TestSender_SendDeadlinedNotification(t *testing.T) {
	if err := boilerplate_testing.FlushScyllaAllTables(nil, session, config.Scylla.Keyspace, nil); err != nil {
		t.Fatal(err)
	}
	if err := boilerplate_testing.FlushPostgresAllTables(config.MasterDb, nil, t); err != nil {
		t.Fatal(err)
	}

	likeTemplate := database.RenderTemplate{
		Id:        "content_like",
		Kind:      "default",
		IsGrouped: true,
	}
	if err := gormDb.Create(&likeTemplate).Error; err != nil {
		t.Fatal(err)
	}

	device := database.Device{
		UserId:    2,
		DeviceId:  "2",
		PushToken: "2",
		Platform:  common.DeviceTypeAndroid,
	}

	if err := gormDb.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "userId"}, {Name: "deviceId"}},
		UpdateAll: true,
	}).Create(&device).Error; err != nil {
		t.Fatal(err)
	}

	baseDate := time.Date(2022, 6, 22, 23, 00, 0, 0, time.UTC)

	renderingVariables := database.RenderingVariables{
		"firstname":          "1",
		"lastname":           "",
		"notificationsCount": "3",
	}
	renderingVariablesMarshalled, err := json.Marshal(renderingVariables)
	if err != nil {
		t.Fatal(err)
	}

	notification1 := scylla.Notification{
		UserId:             2,
		EventType:          likeTemplate.Id,
		EntityId:           1,
		RelatedEntityId:    1,
		CreatedAt:          baseDate,
		NotificationsCount: 3,
		Title:              "Lit.it",
		Body:               "1  and 2 others liked your video",
		Kind:               likeTemplate.Kind,
		RenderingVariables: string(renderingVariablesMarshalled),
		CustomData:         "{}",
		NotificationInfo:   "{}",
	}

	if err = session.Query("update notification set notifications_count = ?, title = ?, body = ?, headline = ?, kind = ?, rendering_variables = ?, "+
		"custom_data = ?, notification_info = ? where user_id = ? and event_type = ? "+
		"and created_at = ? and entity_id = ? and related_entity_id = ?", notification1.NotificationsCount, notification1.Title,
		notification1.Body, notification1.Headline, notification1.Kind, notification1.RenderingVariables,
		notification1.CustomData, notification1.NotificationInfo, notification1.UserId, notification1.EventType,
		notification1.CreatedAt, notification1.EntityId, notification1.RelatedEntityId).Exec(); err != nil {
		t.Fatal(err)
	}

	if err = session.Query("update notification_relation set event_applied = true where user_id = ? and event_type = ? "+
		"and entity_id = ? and related_entity_id = ?", notification1.UserId, notification1.EventType,
		notification1.EntityId, notification1.RelatedEntityId).Exec(); err != nil {
		t.Fatal(err)
	}

	pushNotificationGroupQueue := scylla.PushNotificationGroupQueue{
		DeadlineKey:       TimeToNearestMinutes(baseDate, configs.PushNotificationDeadlineKeyMinutes, false),
		Deadline:          TimeToNearestMinutes(baseDate, configs.PushNotificationDeadlineMinutes, false),
		UserId:            notification1.UserId,
		EventType:         notification1.EventType,
		EntityId:          notification1.EntityId,
		CreatedAt:         notification1.CreatedAt,
		NotificationCount: notification1.NotificationsCount,
	}

	if err = session.Query("update push_notification_group_queue set created_at = ?, notification_count = ? "+
		"where deadline_key = ? and deadline = ? and user_id = ? and event_type = ? and entity_id = ?",
		pushNotificationGroupQueue.CreatedAt, pushNotificationGroupQueue.NotificationCount,
		pushNotificationGroupQueue.DeadlineKey, pushNotificationGroupQueue.Deadline, pushNotificationGroupQueue.UserId,
		pushNotificationGroupQueue.EventType, pushNotificationGroupQueue.EntityId).Exec(); err != nil {
		t.Fatal(err)
	}

	pushSendMessages = nil
	if _, err = sender.SendDeadlinedNotification(pushNotificationGroupQueue.Deadline.Add(-1*time.Second),
		pushNotificationGroupQueue, context.TODO()); err != nil {
		t.Fatal(err)
	}

	iter := session.Query("select deadline_key, deadline, user_id, event_type, entity_id, created_at, "+
		"notification_count from push_notification_group_queue where deadline_key = ? and deadline = ? and user_id = ? "+
		"and event_type = ? and entity_id = ?", pushNotificationGroupQueue.DeadlineKey, pushNotificationGroupQueue.Deadline,
		pushNotificationGroupQueue.UserId, pushNotificationGroupQueue.EventType, pushNotificationGroupQueue.EntityId).Iter()

	pushNotificationGroupQueueUpdated := scylla.PushNotificationGroupQueue{}
	iter.Scan(&pushNotificationGroupQueueUpdated.DeadlineKey, &pushNotificationGroupQueueUpdated.Deadline,
		&pushNotificationGroupQueueUpdated.UserId, &pushNotificationGroupQueueUpdated.EventType,
		&pushNotificationGroupQueueUpdated.EntityId, &pushNotificationGroupQueueUpdated.CreatedAt,
		&pushNotificationGroupQueueUpdated.NotificationCount)

	if err = iter.Close(); err != nil {
		t.Fatal(err)
	}

	a := assert.New(t)

	a.Equal(pushNotificationGroupQueue.DeadlineKey, pushNotificationGroupQueueUpdated.DeadlineKey)
	a.Equal(pushNotificationGroupQueue.Deadline, pushNotificationGroupQueueUpdated.Deadline)
	a.Equal(pushNotificationGroupQueue.UserId, pushNotificationGroupQueueUpdated.UserId)
	a.Equal(pushNotificationGroupQueue.EventType, pushNotificationGroupQueueUpdated.EventType)
	a.Equal(pushNotificationGroupQueue.EntityId, pushNotificationGroupQueueUpdated.EntityId)
	a.Equal(pushNotificationGroupQueue.CreatedAt, pushNotificationGroupQueueUpdated.CreatedAt)
	a.Equal(pushNotificationGroupQueue.NotificationCount, pushNotificationGroupQueueUpdated.NotificationCount)
	a.Len(pushSendMessages, 0)

	pushSendMessages = nil
	if _, err = sender.SendDeadlinedNotification(pushNotificationGroupQueue.Deadline.Add((2*configs.PushNotificationDeadlineKeyMinutes+1)*time.Minute),
		pushNotificationGroupQueue, context.TODO()); err != nil {
		t.Fatal(err)
	}

	iter = session.Query("select deadline_key, deadline, user_id, event_type, entity_id, created_at, "+
		"notification_count from push_notification_group_queue where deadline_key = ? and deadline = ? and user_id = ? "+
		"and event_type = ? and entity_id = ?", pushNotificationGroupQueue.DeadlineKey, pushNotificationGroupQueue.Deadline,
		pushNotificationGroupQueue.UserId, pushNotificationGroupQueue.EventType, pushNotificationGroupQueue.EntityId).Iter()

	pushNotificationGroupQueueUpdated = scylla.PushNotificationGroupQueue{}
	iter.Scan(&pushNotificationGroupQueueUpdated.DeadlineKey, &pushNotificationGroupQueueUpdated.Deadline,
		&pushNotificationGroupQueueUpdated.UserId, &pushNotificationGroupQueueUpdated.EventType,
		&pushNotificationGroupQueueUpdated.EntityId, &pushNotificationGroupQueueUpdated.CreatedAt,
		&pushNotificationGroupQueueUpdated.NotificationCount)

	if err = iter.Close(); err != nil {
		t.Fatal(err)
	}

	a.Equal(pushNotificationGroupQueue.DeadlineKey, pushNotificationGroupQueueUpdated.DeadlineKey)
	a.Equal(pushNotificationGroupQueue.Deadline, pushNotificationGroupQueueUpdated.Deadline)
	a.Equal(pushNotificationGroupQueue.UserId, pushNotificationGroupQueueUpdated.UserId)
	a.Equal(pushNotificationGroupQueue.EventType, pushNotificationGroupQueueUpdated.EventType)
	a.Equal(pushNotificationGroupQueue.EntityId, pushNotificationGroupQueueUpdated.EntityId)
	a.Equal(pushNotificationGroupQueue.CreatedAt, pushNotificationGroupQueueUpdated.CreatedAt)
	a.Equal(pushNotificationGroupQueue.NotificationCount, pushNotificationGroupQueueUpdated.NotificationCount)
	a.Len(pushSendMessages, 0)

	pushSendMessages = nil
	if _, err = sender.SendDeadlinedNotification(pushNotificationGroupQueue.Deadline.Add(1*time.Second),
		pushNotificationGroupQueue, context.TODO()); err != nil {
		t.Fatal(err)
	}

	iter = session.Query("select deadline_key, deadline, user_id, event_type, entity_id, created_at, "+
		"notification_count from push_notification_group_queue where deadline_key = ? and deadline = ? and user_id = ? "+
		"and event_type = ? and entity_id = ?", pushNotificationGroupQueue.DeadlineKey, pushNotificationGroupQueue.Deadline,
		pushNotificationGroupQueue.UserId, pushNotificationGroupQueue.EventType, pushNotificationGroupQueue.EntityId).Iter()

	pushNotificationGroupQueueUpdated = scylla.PushNotificationGroupQueue{}
	iter.Scan(&pushNotificationGroupQueueUpdated.DeadlineKey, &pushNotificationGroupQueueUpdated.Deadline,
		&pushNotificationGroupQueueUpdated.UserId, &pushNotificationGroupQueueUpdated.EventType,
		&pushNotificationGroupQueueUpdated.EntityId, &pushNotificationGroupQueueUpdated.CreatedAt,
		&pushNotificationGroupQueueUpdated.NotificationCount)

	if err = iter.Close(); err != nil {
		t.Fatal(err)
	}

	a.Equal(int64(0), pushNotificationGroupQueueUpdated.UserId)
	a.Len(pushSendMessages, 1)
	a.Equal(notification1.Title, pushSendMessages[0].Title)
	a.Equal(notification1.Body, pushSendMessages[0].Body)

	pushNotificationGroupQueue = scylla.PushNotificationGroupQueue{
		DeadlineKey:       TimeToNearestMinutes(baseDate, configs.PushNotificationDeadlineKeyMinutes, false),
		Deadline:          TimeToNearestMinutes(baseDate, configs.PushNotificationDeadlineMinutes, false),
		UserId:            notification1.UserId,
		EventType:         notification1.EventType,
		EntityId:          notification1.EntityId + 1,
		CreatedAt:         notification1.CreatedAt,
		NotificationCount: notification1.NotificationsCount,
	}

	if err = session.Query("update push_notification_group_queue set created_at = ?, notification_count = ? "+
		"where deadline_key = ? and deadline = ? and user_id = ? and event_type = ? and entity_id = ?",
		pushNotificationGroupQueue.CreatedAt, pushNotificationGroupQueue.NotificationCount,
		pushNotificationGroupQueue.DeadlineKey, pushNotificationGroupQueue.Deadline, pushNotificationGroupQueue.UserId,
		pushNotificationGroupQueue.EventType, pushNotificationGroupQueue.EntityId).Exec(); err != nil {
		t.Fatal(err)
	}

	pushSendMessages = nil
	if _, err = sender.SendDeadlinedNotification(pushNotificationGroupQueue.Deadline.Add(1*time.Second),
		pushNotificationGroupQueue, context.TODO()); err != nil {
		t.Fatal(err)
	}

	iter = session.Query("select deadline_key, deadline, user_id, event_type, entity_id, created_at, "+
		"notification_count from push_notification_group_queue where deadline_key = ? and deadline = ? and user_id = ? "+
		"and event_type = ? and entity_id = ?", pushNotificationGroupQueue.DeadlineKey, pushNotificationGroupQueue.Deadline,
		pushNotificationGroupQueue.UserId, pushNotificationGroupQueue.EventType, pushNotificationGroupQueue.EntityId).Iter()

	pushNotificationGroupQueueUpdated = scylla.PushNotificationGroupQueue{}
	iter.Scan(&pushNotificationGroupQueueUpdated.DeadlineKey, &pushNotificationGroupQueueUpdated.Deadline,
		&pushNotificationGroupQueueUpdated.UserId, &pushNotificationGroupQueueUpdated.EventType,
		&pushNotificationGroupQueueUpdated.EntityId, &pushNotificationGroupQueueUpdated.CreatedAt,
		&pushNotificationGroupQueueUpdated.NotificationCount)

	if err = iter.Close(); err != nil {
		t.Fatal(err)
	}

	a.Equal(pushNotificationGroupQueue.DeadlineKey, pushNotificationGroupQueueUpdated.DeadlineKey)
	a.Equal(pushNotificationGroupQueue.Deadline, pushNotificationGroupQueueUpdated.Deadline)
	a.Equal(pushNotificationGroupQueue.UserId, pushNotificationGroupQueueUpdated.UserId)
	a.Equal(pushNotificationGroupQueue.EventType, pushNotificationGroupQueueUpdated.EventType)
	a.Equal(pushNotificationGroupQueue.EntityId, pushNotificationGroupQueueUpdated.EntityId)
	a.Equal(pushNotificationGroupQueue.CreatedAt, pushNotificationGroupQueueUpdated.CreatedAt)
	a.Equal(pushNotificationGroupQueue.NotificationCount, pushNotificationGroupQueueUpdated.NotificationCount)
	a.Len(pushSendMessages, 0)

	pushNotificationGroupQueue = scylla.PushNotificationGroupQueue{
		DeadlineKey:       TimeToNearestMinutes(baseDate, configs.PushNotificationDeadlineKeyMinutes, false),
		Deadline:          TimeToNearestMinutes(baseDate, configs.PushNotificationDeadlineMinutes, false),
		UserId:            notification1.UserId,
		EventType:         notification1.EventType,
		EntityId:          notification1.EntityId,
		CreatedAt:         notification1.CreatedAt,
		NotificationCount: notification1.NotificationsCount,
	}

	if err = session.Query("update push_notification_group_queue set created_at = ?, notification_count = ? "+
		"where deadline_key = ? and deadline = ? and user_id = ? and event_type = ? and entity_id = ?",
		pushNotificationGroupQueue.CreatedAt, pushNotificationGroupQueue.NotificationCount,
		pushNotificationGroupQueue.DeadlineKey, pushNotificationGroupQueue.Deadline, pushNotificationGroupQueue.UserId,
		pushNotificationGroupQueue.EventType, pushNotificationGroupQueue.EntityId).Exec(); err != nil {
		t.Fatal(err)
	}

	pushSendMessages = nil
	if _, err = sender.SendDeadlinedNotification(pushNotificationGroupQueue.Deadline,
		pushNotificationGroupQueue, context.TODO()); err != nil {
		t.Fatal(err)
	}

	iter = session.Query("select deadline_key, deadline, user_id, event_type, entity_id, created_at, "+
		"notification_count from push_notification_group_queue where deadline_key = ? and deadline = ? and user_id = ? "+
		"and event_type = ? and entity_id = ?", pushNotificationGroupQueue.DeadlineKey, pushNotificationGroupQueue.Deadline,
		pushNotificationGroupQueue.UserId, pushNotificationGroupQueue.EventType, pushNotificationGroupQueue.EntityId).Iter()

	pushNotificationGroupQueueUpdated = scylla.PushNotificationGroupQueue{}
	iter.Scan(&pushNotificationGroupQueueUpdated.DeadlineKey, &pushNotificationGroupQueueUpdated.Deadline,
		&pushNotificationGroupQueueUpdated.UserId, &pushNotificationGroupQueueUpdated.EventType,
		&pushNotificationGroupQueueUpdated.EntityId, &pushNotificationGroupQueueUpdated.CreatedAt,
		&pushNotificationGroupQueueUpdated.NotificationCount)

	if err = iter.Close(); err != nil {
		t.Fatal(err)
	}

	a.Equal(int64(0), pushNotificationGroupQueueUpdated.UserId)
	a.Len(pushSendMessages, 1)
	a.Equal(notification1.Title, pushSendMessages[0].Title)
	a.Equal(notification1.Body, pushSendMessages[0].Body)
}
