package sender

import (
	"context"
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/digitalmonsters/go-common/translation"
	"github.com/digitalmonsters/notification-handler/configs"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/database/scylla"
	"github.com/gocql/gocql"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"os"
	"testing"
)

var config configs.Settings
var gormDb *gorm.DB
var session *gocql.Session
var sender *Sender

func TestMain(m *testing.M) {
	config = configs.GetConfig()
	session = database.GetScyllaSession()
	gormDb = database.GetDb(database.DbTypeMaster)
	sender = NewSender(nil, nil, nil)

	os.Exit(m.Run())
}

func TestSender_PushNotification(t *testing.T) {
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
			"firstname": "test_f1",
			"lastname":  "test_l1",
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
			"firstname": "test_f1",
			"lastname":  "test_l1",
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
