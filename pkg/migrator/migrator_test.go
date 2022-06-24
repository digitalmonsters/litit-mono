package migrator

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/digitalmonsters/go-common/wrappers"
	"github.com/digitalmonsters/go-common/wrappers/follow"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/notification-handler/configs"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/database/scylla"
	"github.com/digitalmonsters/notification-handler/pkg/notification"
	"github.com/digitalmonsters/notification-handler/pkg/utils"
	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

var config configs.Settings
var gormDb *gorm.DB
var session *gocql.Session
var userWrapperMock user_go.IUserGoWrapper
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

func TestMigrateNotificationsToScyllaWithSeed(t *testing.T) {
	if err := boilerplate_testing.FlushScyllaAllTables(nil, session, config.Scylla.Keyspace, nil); err != nil {
		t.Fatal(err)
	}
	if err := boilerplate_testing.FlushPostgresAllTables(config.MasterDb, nil, t); err != nil {
		t.Fatal(err)
	}

	if err := utils.PollutePostgresDatabase(gormDb, "./test_data/seed.json"); err != nil {
		t.Fatal(err)
	}

	if err := utils.PollutePostgresDatabase(gormDb, "./test_data/templates.json"); err != nil {
		t.Fatal(err)
	}

	if err := MigrateNotificationsToScylla(context.TODO()); err != nil {
		t.Fatal(err)
	}

	resp, err := notification.GetNotifications(gormDb, 1074760, "", notification.TypeGroupAll, 20,
		userWrapperMock, followWrapper, context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	a := assert.New(t)

	a.NotNil(resp)
	a.Len(resp.Data, 12)

	path, err := boilerplate.RecursiveFindFile("./test_data/expected_seed.json", "./", 30)
	if err != nil {
		t.Fatal(err)
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var dataExpected []notification.NotificationsResponseItem

	if err = json.Unmarshal(data, &dataExpected); err != nil {
		t.Fatal(err)
	}

	for _, expected := range dataExpected {
		foundItem, found := lo.Find(resp.Data, func(item notification.NotificationsResponseItem) bool {
			return item.Id == expected.Id
		})

		a.True(found)
		a.Equal(expected.Id, foundItem.Id)
		a.Equal(expected.UserId, foundItem.UserId)
		a.Equal(expected.Type, foundItem.Type)
		a.Equal(expected.Title, foundItem.Title)
		a.Equal(expected.Message, foundItem.Message)
		a.Equal(expected.RelatedUserId, foundItem.RelatedUserId)
		a.Equal(expected.RelatedUser, foundItem.RelatedUser)
		a.Equal(expected.RenderingVariables, foundItem.RenderingVariables)
		a.Equal(expected.CustomData, foundItem.CustomData)
		a.Equal(expected.CommentId, foundItem.CommentId)
		a.Equal(expected.Comment, foundItem.Comment)
		a.Equal(expected.ContentId, foundItem.ContentId)
		a.Equal(expected.Content, foundItem.Content)
		a.Equal(expected.QuestionId, foundItem.QuestionId)
		a.Equal(expected.KycStatus, foundItem.KycStatus)
		a.Equal(expected.ContentCreatorStatus, foundItem.ContentCreatorStatus)
		a.Equal(expected.KycReason, foundItem.KycReason)
		a.Equal(expected.CreatedAt.UTC(), foundItem.CreatedAt.UTC())
		a.Equal(expected.NotificationsCount, foundItem.NotificationsCount)
	}
}
