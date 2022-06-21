package migrator

import (
	"context"
	"encoding/json"
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/digitalmonsters/notification-handler/configs"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/database/scylla"
	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"os"
	"testing"
	"time"
)

var config configs.Settings
var gormDb *gorm.DB
var session *gocql.Session

func TestMain(m *testing.M) {
	config = configs.GetConfig()
	session = database.GetScyllaSession()
	gormDb = database.GetDb(database.DbTypeMaster)

	os.Exit(m.Run())
}

func TestMigrateNotificationsToScylla(t *testing.T) {
	if err := boilerplate_testing.FlushScyllaAllTables(nil, session, config.Scylla.Keyspace, nil); err != nil {
		t.Fatal(err)
	}
	if err := boilerplate_testing.FlushPostgresAllTables(config.MasterDb, nil, t); err != nil {
		t.Fatal(err)
	}

	if err := gormDb.Create(&database.RenderTemplate{
		Id:        "comment_vote_like",
		Kind:      "default",
		IsGrouped: true,
	}).Error; err != nil {
		t.Fatal(err)
	}

	if err := gormDb.Create(&database.RenderTemplate{
		Id:        "comment_vote_dislike",
		Kind:      "default",
		IsGrouped: true,
	}).Error; err != nil {
		t.Fatal(err)
	}

	if err := gormDb.Create(&database.RenderTemplate{
		Id:        "bonus_time",
		Kind:      "default",
		IsGrouped: false,
	}).Error; err != nil {
		t.Fatal(err)
	}

	baseDate := time.Date(2022, 6, 17, 10, 25, 0, 0, time.UTC)

	dbNotification0 := database.Notification{
		Id:            uuid.New(),
		UserId:        2,
		Type:          "push.comment.vote",
		Title:         "1",
		Message:       "100 liked your comment: 1",
		RelatedUserId: null.IntFrom(100),
		CommentId:     null.IntFrom(1),
		Comment: &database.NotificationComment{
			Id:      1,
			Type:    1,
			Comment: "1",
		},
		CreatedAt: baseDate.Add(-3 * 24 * 100 * time.Hour),
		RenderingVariables: database.RenderingVariables{
			"firstname": "100",
			"lastname":  "",
			"comment":   "1",
		},
	}
	if err := gormDb.Create(&dbNotification0).Error; err != nil {
		t.Fatal(err)
	}

	dbNotification1 := database.Notification{
		Id:            uuid.New(),
		UserId:        2,
		Type:          "push.comment.vote",
		Title:         "1",
		Message:       "1 liked your comment: 1",
		RelatedUserId: null.IntFrom(1),
		CommentId:     null.IntFrom(1),
		Comment: &database.NotificationComment{
			Id:      1,
			Type:    1,
			Comment: "1",
		},
		CreatedAt: baseDate,
		RenderingVariables: database.RenderingVariables{
			"firstname": "1",
			"lastname":  "",
			"comment":   "1",
		},
	}
	if err := gormDb.Create(&dbNotification1).Error; err != nil {
		t.Fatal(err)
	}

	dbNotification2 := database.Notification{
		Id:            uuid.New(),
		UserId:        2,
		Type:          "push.comment.vote",
		Title:         "21",
		Message:       "21 liked your comment: 1",
		RelatedUserId: null.IntFrom(21),
		CommentId:     null.IntFrom(1),
		Comment: &database.NotificationComment{
			Id:      1,
			Type:    1,
			Comment: "1",
		},
		CreatedAt: baseDate.Add(1 * time.Minute),
		RenderingVariables: database.RenderingVariables{
			"firstname": "21",
			"lastname":  "",
			"comment":   "1",
		},
	}
	if err := gormDb.Create(&dbNotification2).Error; err != nil {
		t.Fatal(err)
	}

	dbNotification3 := database.Notification{
		Id:            uuid.New(),
		UserId:        2,
		Type:          "push.comment.vote",
		Title:         "Lit.it",
		Message:       "31 liked your comment: 1",
		RelatedUserId: null.IntFrom(31),
		CommentId:     null.IntFrom(1),
		Comment: &database.NotificationComment{
			Id:      1,
			Type:    1,
			Comment: "1",
		},
		CreatedAt: time.Now().UTC().Add(5 * time.Minute),
		RenderingVariables: database.RenderingVariables{
			"firstname": "31",
			"lastname":  "",
			"comment":   "1",
		},
	}
	if err := gormDb.Create(&dbNotification3).Error; err != nil {
		t.Fatal(err)
	}

	dbNotification3RenderingVariablesMarshalled, err := json.Marshal(dbNotification3.RenderingVariables)
	if err != nil {
		t.Fatal(err)
	}

	dbNotification3Marshalled, err := json.Marshal(dbNotification3)
	if err != nil {
		t.Fatal(err)
	}

	scyllaNotification1 := scylla.Notification{
		UserId:             dbNotification3.UserId,
		EventType:          "comment_vote_like",
		EntityId:           dbNotification3.CommentId.Int64,
		RelatedEntityId:    dbNotification3.RelatedUserId.Int64,
		CreatedAt:          dbNotification3.CreatedAt,
		NotificationsCount: 1,
		Title:              dbNotification3.Title,
		Body:               dbNotification3.Message,
		Kind:               "default",
		RenderingVariables: string(dbNotification3RenderingVariablesMarshalled),
		CustomData:         "{}",
		NotificationInfo:   string(dbNotification3Marshalled),
	}
	if err = session.Query("update notification set notifications_count = ?, title = ?, body = ?, headline = ?, kind = ?, rendering_variables = ?, "+
		"custom_data = ?, notification_info = ? where user_id = ? and event_type = ? "+
		"and created_at = ? and entity_id = ? and related_entity_id = ?", scyllaNotification1.NotificationsCount, scyllaNotification1.Title, scyllaNotification1.Body, scyllaNotification1.Headline,
		scyllaNotification1.Kind, scyllaNotification1.RenderingVariables, scyllaNotification1.CustomData, scyllaNotification1.NotificationInfo,
		scyllaNotification1.UserId, scyllaNotification1.EventType, scyllaNotification1.CreatedAt, scyllaNotification1.EntityId, scyllaNotification1.RelatedEntityId).Exec(); err != nil {
		t.Fatal(err)
	}

	if err = session.Query("update notification_relation set event_applied = true where user_id = ? and event_type = ? "+
		"and entity_id = ? and related_entity_id = ?", scyllaNotification1.UserId, scyllaNotification1.EventType,
		scyllaNotification1.EntityId, scyllaNotification1.RelatedEntityId).Exec(); err != nil {
		t.Fatal(err)
	}

	dbNotification4 := database.Notification{
		Id:            uuid.New(),
		UserId:        2,
		Type:          "push.bonus.daily",
		Title:         "Lit.it",
		Message:       "Daily reward for views",
		RelatedUserId: null.IntFrom(2),
		CreatedAt:     baseDate.Add(2 * time.Minute),
	}
	if err = gormDb.Create(&dbNotification4).Error; err != nil {
		t.Fatal(err)
	}

	dbNotification5 := database.Notification{
		Id:            uuid.New(),
		UserId:        3,
		Type:          "push.comment.vote",
		Title:         "Lit.it",
		Message:       "1  disliked your comment: 2",
		RelatedUserId: null.IntFrom(1),
		CommentId:     null.IntFrom(2),
		Comment: &database.NotificationComment{
			Id:      2,
			Type:    1,
			Comment: "2",
		},
		CreatedAt: baseDate,
		RenderingVariables: database.RenderingVariables{
			"firstname": "1",
			"lastname":  "",
			"comment":   "2",
		},
	}
	if err = gormDb.Create(&dbNotification5).Error; err != nil {
		t.Fatal(err)
	}

	if err = MigrateNotificationsToScylla(context.TODO()); err != nil {
		t.Fatal(err)
	}

	iter := session.Query("select user_id, event_type, entity_id, related_entity_id, created_at, notifications_count, " +
		"title, body, headline, kind, rendering_variables, custom_data, notification_info from notification").Iter()

	var notification scylla.Notification
	notifications := make(map[int64]map[string][]scylla.Notification)
	for iter.Scan(&notification.UserId, &notification.EventType, &notification.EntityId, &notification.RelatedEntityId,
		&notification.CreatedAt, &notification.NotificationsCount, &notification.Title, &notification.Body,
		&notification.Headline, &notification.Kind, &notification.RenderingVariables, &notification.CustomData, &notification.NotificationInfo) {
		userNotifications, ok := notifications[notification.UserId]
		if !ok {
			notifications[notification.UserId] = map[string][]scylla.Notification{
				notification.EventType: {notification},
			}
		} else {
			notificationByType, ok := userNotifications[notification.EventType]
			if !ok {
				notifications[notification.UserId][notification.EventType] = []scylla.Notification{notification}
			} else {
				notifications[notification.UserId][notification.EventType] = append(notificationByType, notification)
			}
		}
	}

	if err = iter.Close(); err != nil {
		t.Fatal(err)
	}

	a := assert.New(t)

	a.Len(notifications, 2)

	scyllaNotifications, ok := notifications[dbNotification1.UserId]["comment_vote_like"]
	a.True(ok)
	a.Len(scyllaNotifications, 1)
	a.Equal(scyllaNotification1.UserId, scyllaNotifications[0].UserId)
	a.Equal(scyllaNotification1.EventType, scyllaNotifications[0].EventType)
	a.Equal(scyllaNotification1.EntityId, scyllaNotifications[0].EntityId)
	a.Equal(scyllaNotification1.RelatedEntityId, scyllaNotifications[0].RelatedEntityId)
	a.Equal(int64(3), scyllaNotifications[0].NotificationsCount)
	a.Equal(scyllaNotification1.Title, scyllaNotifications[0].Title)
	a.Equal("31  and 3 others liked your comment: 1", scyllaNotifications[0].Body)
	a.Equal(scyllaNotification1.Kind, scyllaNotifications[0].Kind)
	a.Equal(scyllaNotification1.CustomData, scyllaNotifications[0].CustomData)
	a.Equal(scyllaNotification1.NotificationInfo, scyllaNotifications[0].NotificationInfo)

	scyllaNotifications, ok = notifications[dbNotification5.UserId]["comment_vote_dislike"]
	a.True(ok)
	a.Len(scyllaNotifications, 1)
	a.Equal(dbNotification5.UserId, scyllaNotifications[0].UserId)
	a.Equal("comment_vote_dislike", scyllaNotifications[0].EventType)
	a.Equal(dbNotification5.CommentId.Int64, scyllaNotifications[0].EntityId)
	a.Equal(dbNotification5.RelatedUserId.Int64, scyllaNotifications[0].RelatedEntityId)
	a.Equal(dbNotification5.CreatedAt, scyllaNotifications[0].CreatedAt)
	a.Equal(int64(1), scyllaNotifications[0].NotificationsCount)
	a.Equal(dbNotification5.Title, scyllaNotifications[0].Title)
	a.Equal(dbNotification5.Message, scyllaNotifications[0].Body)
	a.Equal("default", scyllaNotifications[0].Kind)

	scyllaNotifications, ok = notifications[dbNotification4.UserId]["bonus_time"]
	a.True(ok)
	a.Len(scyllaNotifications, 1)
	a.Equal(dbNotification4.UserId, scyllaNotifications[0].UserId)
	a.Equal("bonus_time", scyllaNotifications[0].EventType)
	a.Equal(dbNotification4.UserId, scyllaNotifications[0].EntityId)
	a.Equal(int64(0), scyllaNotifications[0].RelatedEntityId)
	a.Equal(dbNotification4.CreatedAt, scyllaNotifications[0].CreatedAt)
	a.Equal(int64(1), scyllaNotifications[0].NotificationsCount)
	a.Equal(dbNotification4.Title, scyllaNotifications[0].Title)
	a.Equal(dbNotification4.Message, scyllaNotifications[0].Body)
	a.Equal("default", scyllaNotifications[0].Kind)

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

	a.Equal(4, len(notificationRelations))

	_, ok = lo.Find(notificationRelations, func(item scylla.NotificationRelation) bool {
		nt := dbNotification1
		return nt.UserId == item.UserId && item.EventType == "comment_vote_like" && nt.CommentId.Int64 == item.EntityId &&
			nt.RelatedUserId.Int64 == item.RelatedEntityId && item.EventApplied
	})
	a.True(ok)

	_, ok = lo.Find(notificationRelations, func(item scylla.NotificationRelation) bool {
		nt := dbNotification2
		return nt.UserId == item.UserId && item.EventType == "comment_vote_like" && nt.CommentId.Int64 == item.EntityId &&
			nt.RelatedUserId.Int64 == item.RelatedEntityId && item.EventApplied
	})
	a.True(ok)

	_, ok = lo.Find(notificationRelations, func(item scylla.NotificationRelation) bool {
		nt := dbNotification3
		return nt.UserId == item.UserId && item.EventType == "comment_vote_like" && nt.CommentId.Int64 == item.EntityId &&
			nt.RelatedUserId.Int64 == item.RelatedEntityId && item.EventApplied
	})
	a.True(ok)

	_, ok = lo.Find(notificationRelations, func(item scylla.NotificationRelation) bool {
		nt := dbNotification5
		return nt.UserId == item.UserId && item.EventType == "comment_vote_dislike" && nt.CommentId.Int64 == item.EntityId &&
			nt.RelatedUserId.Int64 == item.RelatedEntityId && item.EventApplied
	})
	a.True(ok)
}
