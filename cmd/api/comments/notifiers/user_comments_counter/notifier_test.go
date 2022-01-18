package user_comments_counter

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/comments/configs"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/gocql/gocql"
	"go.elastic.co/apm"
	"os"
	"reflect"
	"testing"
	"time"
)

var cfg *configs.Settings
var service *Notifier
var cluster *gocql.ClusterConfig
var kafkaPublishedEvents []eventsourcing.IEventData
var pollTime time.Duration
var publisherMock KafkaEventPublisherMock

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

	cfg = &config
	kafkaPublishedEvents = nil

	pollTime = 100 * time.Millisecond

	publisherMock = KafkaEventPublisherMock{}

	service = NewNotifier(
		pollTime,
		context.TODO(),
		&publisherMock,
		false,
	)

	os.Exit(m.Run())
}

func testInsert(t *testing.T) {
	dict := map[int64]int64{
		1: 1,
		2: 2,
		3: 3,
		4: 4,
		5: 5,
	}

	for userId, commentsCount := range dict {
		service.Enqueue(userId, commentsCount)
	}

	service.Flush()

	var newDict = make(map[int64]int64, len(kafkaPublishedEvents))

	i := 0
	for _, event := range kafkaPublishedEvents {
		if i < len(dict) {
			elem := reflect.ValueOf(&event).Elem().Elem()
			id := elem.FieldByName("UserId").Int()
			newDict[id] = elem.FieldByName("Count").Int()
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

	s := NewNotifier(
		pollTime,
		context.TODO(),
		&publisherMock,
		false,
	)

	for i := int64(0); i < 100000; i++ {
		s.Enqueue(i, i)
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
