package vote

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/comments/configs"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
	"os"
	"reflect"
	"sort"
	"testing"
	"time"
)

var cfg *configs.Settings
var service *Notifier
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
	dict := map[string]eventData{
		"1_1": {
			UserId:          1,
			Upvote:          null.BoolFrom(true),
			CommentId:       1,
			ParentId:        null.Int{},
			CommentAuthorId: 1,
			Comment:         "1",
			ContentId:       null.IntFrom(1),
			ProfileId:       null.Int{},
		},
		"2_2": {
			UserId:          2,
			Upvote:          null.BoolFrom(true),
			CommentId:       2,
			ParentId:        null.Int{},
			CommentAuthorId: 2,
			Comment:         "2",
			ContentId:       null.IntFrom(2),
			ProfileId:       null.Int{},
		},
		"3_3": {
			UserId:          3,
			Upvote:          null.BoolFrom(true),
			CommentId:       3,
			ParentId:        null.Int{},
			CommentAuthorId: 3,
			Comment:         "3",
			ContentId:       null.IntFrom(3),
			ProfileId:       null.Int{},
		},
		"4_4": {
			UserId:          4,
			Upvote:          null.BoolFrom(true),
			CommentId:       4,
			ParentId:        null.Int{},
			CommentAuthorId: 4,
			Comment:         "4",
			ContentId:       null.IntFrom(4),
			ProfileId:       null.Int{},
		},
		"5_5": {
			UserId:          5,
			Upvote:          null.BoolFrom(true),
			CommentId:       5,
			ParentId:        null.Int{},
			CommentAuthorId: 5,
			Comment:         "5",
			ContentId:       null.IntFrom(5),
			ProfileId:       null.Int{},
		},
	}

	for _, event := range dict {
		service.Enqueue(event.CommentId, event.UserId, event.Upvote, event.ParentId, event.CommentAuthorId,
			event.Comment, event.ContentId, event.ProfileId)
	}

	errs := service.Flush()
	assert.Equal(t, len(errs), 0)
	for _, err := range errs {
		log.Err(err).Send()
	}

	var newDict = make(map[string]eventData, len(kafkaPublishedEvents))
	sort.Slice(kafkaPublishedEvents, func(i, j int) bool {
		return kafkaPublishedEvents[i].(eventData).UserId < kafkaPublishedEvents[j].(eventData).UserId
	})

	sort.Slice(kafkaPublishedEvents, func(i, j int) bool {
		return kafkaPublishedEvents[i].(eventData).UserId < kafkaPublishedEvents[j].(eventData).UserId
	})

	i := 0
	for _, event := range kafkaPublishedEvents {
		if i < len(dict) {
			elem := reflect.ValueOf(&event).Elem().Elem()
			commentId := elem.FieldByName("CommentId").Int()
			userId := elem.FieldByName("UserId").Int()
			newDict[fmt.Sprintf("%v_%v", commentId, userId)] = eventData{
				CommentId:       commentId,
				UserId:          userId,
				Upvote:          elem.FieldByName("Upvote").Interface().(null.Bool),
				ParentId:        elem.FieldByName("ParentId").Interface().(null.Int),
				CommentAuthorId: elem.FieldByName("CommentAuthorId").Int(),
				ContentId:       elem.FieldByName("ContentId").Interface().(null.Int),
				Comment:         elem.FieldByName("Comment").String(),
				ProfileId:       elem.FieldByName("ProfileId").Interface().(null.Int),
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

	s := NewNotifier(
		pollTime,
		context.TODO(),
		&publisherMock,
		false,
	)

	for i := int64(0); i < 100000; i++ {
		s.Enqueue(i, i, null.BoolFrom(true), null.IntFrom(i), i, fmt.Sprint(i), null.IntFrom(i), null.IntFrom(i))
	}

	b.ResetTimer()

	start := time.Now()

	s.Flush()

	duration := time.Since(start)

	fmt.Printf("Per second : %v", float64(contentCount)/duration.Seconds())
	fmt.Printf("Duration in seconds: %v", duration.Seconds())
}
