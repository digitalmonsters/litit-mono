package notification

import (
	"context"
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/digitalmonsters/notification-handler/configs"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/gocql/gocql"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var config configs.Settings
var session *gocql.Session

func TestMain(m *testing.M) {
	config = configs.GetConfig()
	session = database.GetScyllaSession()

	os.Exit(m.Run())
}

func TestReadNotification(t *testing.T) {
	if err := boilerplate_testing.FlushScyllaAllTables(nil, session, config.Scylla.Keyspace, nil); err != nil {
		t.Fatal(err)
	}
	if err := boilerplate_testing.FlushPostgresAllTables(config.MasterDb, nil, t); err != nil {
		t.Fatal(err)
	}

	req := ReadNotificationRequest{
		NotificationId: 1,
	}
	userId := int64(1)
	err := ReadNotification(req, userId, context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	a := assert.New(t)
	iter := session.Query("select notification_id from user_notifications_read where cluster_key = ? and notification_id = ? and user_id = ?;",
		GetUserNotificationsReadClusterKey(userId), req.NotificationId, userId).
		WithContext(context.TODO()).Iter()
	if err = iter.Close(); err != nil {
		t.Fatal(err)
	}
	a.Equal(1, iter.NumRows())

	err = ReadNotification(req, userId, context.TODO())
	if err != nil {
		t.Fatal(err)
	}
	iter = session.Query("select notification_id from user_notifications_read where cluster_key = ? and notification_id = ? and user_id = ?;", GetUserNotificationsReadClusterKey(userId), req.NotificationId, userId).WithContext(context.TODO()).Iter()
	a.Equal(1, iter.NumRows())

	iter = session.Query("select read_count from user_notifications_read_counter where notification_id in ?;", []int64{1}).WithContext(context.TODO()).Iter()
	var readCount int64
	for iter.Scan(&readCount) {
		a.Equal(int64(1), readCount)
	}
	if err = iter.Close(); err != nil {
		t.Fatal(err)
	}

	err = ReadNotification(req, 2, context.TODO())
	if err != nil {
		t.Fatal(err)
	}
	iter = session.Query("select read_count from user_notifications_read_counter where notification_id in ?;", []int64{1}).WithContext(context.TODO()).Iter()
	for iter.Scan(&readCount) {
		a.Equal(int64(2), readCount)
	}
	if err = iter.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestGetNotificationsReadCount(t *testing.T) {
	if err := boilerplate_testing.FlushScyllaAllTables(nil, session, config.Scylla.Keyspace, nil); err != nil {
		t.Fatal(err)
	}
	if err := boilerplate_testing.FlushPostgresAllTables(config.MasterDb, nil, t); err != nil {
		t.Fatal(err)
	}

	for i := int64(1); i <= 2; i++ {
		if err := session.Query(
			"update user_notifications_read_counter set read_count = read_count + ? where notification_id = ?",
			11, i,
		).Exec(); err != nil {
			t.Fatal(err)
		}
	}

	req := GetNotificationsReadCountRequest{
		NotificationIds: []int64{1, 2},
	}
	resp, err := GetNotificationsReadCount(req, context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	a := assert.New(t)
	a.Equal(int64(11), resp[1])
	a.Equal(int64(11), resp[2])
}
