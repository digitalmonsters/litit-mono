package comment

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/comments/configs"
	"github.com/digitalmonsters/comments/pkg/database"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/gocql/gocql"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"os"
	"reflect"
	"sort"
	"testing"
	"time"
)

var cfg *configs.Settings
var service *Notifier
var cluster *gocql.ClusterConfig
var kafkaPublishedEvents []eventsourcing.IEventData
var pollTime time.Duration
var publisherMock KafkaEventPublisherMock
var gormDb *gorm.DB

type KafkaEventPublisherMock struct {
}

func (s *KafkaEventPublisherMock) GetPublisherType() eventsourcing.PublisherType {
	panic("implement me")
}

func (s *KafkaEventPublisherMock) Publish(apmTransaction *apm.Transaction, events ...eventsourcing.IEventData) []error {
	kafkaPublishedEvents = events
	return nil
}

func TestMain(m *testing.M) {
	config := configs.GetConfig()
	gormDb = database.GetDb()
	cfg = &config
	kafkaPublishedEvents = nil

	pollTime = 100 * time.Millisecond

	publisherMock = KafkaEventPublisherMock{}

	service = NewNotifier(
		pollTime,
		context.TODO(),
		&publisherMock,
		gormDb,
	)

	os.Exit(m.Run())
}

func testInsert(t *testing.T) {
	dict := map[string]database.Comment{
		"1": {
			Id:           1,
			AuthorId:     1,
			NumReplies:   1,
			NumUpvotes:   1,
			NumDownvotes: 1,
			CreatedAt:    time.Now().UTC(),
			Active:       true,
			Comment:      "1",
			ContentId:    null.IntFrom(1),
			ParentId:     null.IntFrom(1),
		},
		"2": {
			Id:           2,
			AuthorId:     2,
			NumReplies:   2,
			NumUpvotes:   2,
			NumDownvotes: 2,
			CreatedAt:    time.Now().UTC(),
			Active:       true,
			Comment:      "2",
			ContentId:    null.IntFrom(1),
			ParentId:     null.IntFrom(2),
		},
		"3": {
			Id:           3,
			AuthorId:     3,
			NumReplies:   3,
			NumUpvotes:   3,
			NumDownvotes: 3,
			CreatedAt:    time.Now().UTC(),
			Active:       true,
			Comment:      "3",
			ContentId:    null.IntFrom(1),
			ParentId:     null.IntFrom(3),
		},
		"4": {
			Id:           4,
			AuthorId:     4,
			NumReplies:   4,
			NumUpvotes:   4,
			NumDownvotes: 4,
			CreatedAt:    time.Now().UTC(),
			Active:       true,
			Comment:      "4",
			ContentId:    null.IntFrom(1),
			ParentId:     null.IntFrom(4),
		},
		"5": {
			Id:           5,
			AuthorId:     5,
			NumReplies:   5,
			NumUpvotes:   5,
			NumDownvotes: 5,
			CreatedAt:    time.Now().UTC(),
			Active:       true,
			Comment:      "5",
			ContentId:    null.IntFrom(1),
			ParentId:     null.IntFrom(5),
		},
	}

	for _, event := range dict {
		service.Enqueue(event, content.SimpleContent{}, ProfileResourceTypeCreate)
	}

	service.Flush()

	var newDict = make(map[string]database.Comment, len(kafkaPublishedEvents))
	sort.Slice(kafkaPublishedEvents, func(i, j int) bool {
		return kafkaPublishedEvents[i].(eventData).Id < kafkaPublishedEvents[j].(eventData).Id
	})

	i := 0
	for _, event := range kafkaPublishedEvents {
		if i < len(dict) {
			elem := reflect.ValueOf(&event).Elem().Elem()
			id := elem.FieldByName("Id").Int()
			newDict[fmt.Sprintf("%v", id)] = database.Comment{
				Id:           id,
				AuthorId:     elem.FieldByName("AuthorId").Int(),
				NumReplies:   elem.FieldByName("NumReplies").Int(),
				NumUpvotes:   elem.FieldByName("NumUpvotes").Int(),
				NumDownvotes: elem.FieldByName("NumDownvotes").Int(),
				CreatedAt:    elem.FieldByName("CreatedAt").Interface().(time.Time),
				Active:       elem.FieldByName("Active").Bool(),
				Comment:      elem.FieldByName("Comment").String(),
				ContentId:    elem.FieldByName("ContentId").Interface().(null.Int),
				ParentId:     elem.FieldByName("ParentId").Interface().(null.Int),
			}
		}
		i++
	}

	if !reflect.DeepEqual(dict, newDict) {
		t.Fatal("Unexpected value")
	}
}

func TestInsert(t *testing.T) {
	testInsert(t)
}

func BenchmarkPerformance(b *testing.B) {
	testPerformance(b)
}

func testPerformance(b *testing.B) {
	var contentCount = int64(100000)

	cfg.NotifierCommentConfig.KafkaTopic = "back_comment_events_test"

	s := NewNotifier(
		pollTime,
		context.TODO(),
		&publisherMock,
		gormDb,
	)

	for i := int64(0); i < 100000; i++ {
		s.Enqueue(database.Comment{
			Id:           i,
			AuthorId:     i,
			NumReplies:   i,
			NumUpvotes:   i,
			NumDownvotes: i,
			CreatedAt:    time.Now().UTC(),
			Active:       true,
			Comment:      fmt.Sprint(i),
			ContentId:    null.IntFrom(i),
			ParentId:     null.IntFrom(i - 1),
		}, content.SimpleContent{}, ProfileResourceTypeCreate)
	}

	b.ResetTimer()

	start := time.Now()

	s.Flush()

	duration := time.Since(start)

	fmt.Printf("Per second : %v", float64(contentCount)/duration.Seconds())
	fmt.Printf("Duration in seconds: %v", duration.Seconds())
	fmt.Printf("Cluster Page Size: %v", cluster.PageSize)
	fmt.Printf("Cluster Connections Number: %v", cluster.NumConns)
	fmt.Printf("Cluster Max Routing Key Info: %v", cluster.MaxRoutingKeyInfo)
	fmt.Printf("Cluster Max Prepared Statements: %v", cluster.MaxPreparedStmts)
}
