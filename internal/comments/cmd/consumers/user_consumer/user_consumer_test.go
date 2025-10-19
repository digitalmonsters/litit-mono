package user_consumer

import (
	"context"
	"github.com/digitalmonsters/comments/configs"
	"github.com/digitalmonsters/comments/pkg/database"
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"os"
	"testing"
)

var db *gorm.DB

func TestMain(m *testing.M) {
	db = database.GetDb()

	os.Exit(m.Run())
}

func TestDeleteChild(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresAllTables(configs.GetConfig().Db, nil, nil); err != nil {
		t.Fatal(err)
	}

	if err := boilerplate_testing.PollutePostgresDatabase(db, "./test_data/seed.json"); err != nil {
		t.Fatal(err)
	}

	_, err := process(wrappedEvent{
		UserEvent: eventsourcing.UserEvent{
			UserId: 1000220,
			BaseChangeEvent: eventsourcing.BaseChangeEvent{
				CrudOperation:       eventsourcing.ChangeEventTypeDeleted,
				CrudOperationReason: eventsourcing.DeleteModeHard,
			},
		},
		Message: kafka.Message{},
	}, nil, context.TODO(), nil, nil, nil)

	if err != nil {
		t.Fatal(err)
	}

	var comment []database.Comment

	if err := db.Order("id asc").Find(&comment).Error; err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 2, len(comment))

	assert.Equal(t, int64(16), comment[0].Id)
	assert.Equal(t, int64(4), comment[0].NumReplies)
	assert.Equal(t, int64(0), comment[0].NumUpvotes)
	assert.Equal(t, int64(1), comment[0].NumDownvotes)

	_, err = process(wrappedEvent{
		UserEvent: eventsourcing.UserEvent{
			UserId: 1000218,
			BaseChangeEvent: eventsourcing.BaseChangeEvent{
				CrudOperation:       eventsourcing.ChangeEventTypeDeleted,
				CrudOperationReason: eventsourcing.DeleteModeHard,
			},
		},
		Message: kafka.Message{},
	}, nil, context.TODO(), nil, nil, nil)

	if err != nil {
		t.Fatal(err)
	}

	if err := db.Order("id asc").Find(&comment).Error; err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 2, len(comment))

	assert.Equal(t, int64(16), comment[0].Id)

	assert.Equal(t, int64(4), comment[0].NumReplies)
	assert.Equal(t, int64(0), comment[0].NumUpvotes)
	assert.Equal(t, int64(0), comment[0].NumDownvotes)

	_, err = process(wrappedEvent{
		UserEvent: eventsourcing.UserEvent{
			UserId: 1000219,
			BaseChangeEvent: eventsourcing.BaseChangeEvent{
				CrudOperation:       eventsourcing.ChangeEventTypeCreated,
				CrudOperationReason: eventsourcing.DeleteModeSoft,
			},
		},
		Message: kafka.Message{},
	}, nil, context.TODO(), nil, nil, nil)

	if err != nil {
		t.Fatal(err)
	}

	if err := db.Order("id asc").Find(&comment).Error; err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 2, len(comment))
}
