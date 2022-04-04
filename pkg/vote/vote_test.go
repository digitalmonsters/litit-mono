package vote

import (
	"context"
	"github.com/digitalmonsters/go-common/wrappers"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/digitalmonsters/comments/cmd/api/comments/notifiers/comment"
	"github.com/digitalmonsters/comments/configs"
	"github.com/digitalmonsters/comments/pkg/database"
	"github.com/digitalmonsters/comments/utils"
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/stretchr/testify/assert"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
)

var db *gorm.DB
var notifier *comment.Notifier
var kafkaPublishedEvents []eventsourcing.IEventData
var pollTime time.Duration
var publisherMock KafkaEventPublisherMock
var contentWrapperMock content.IContentWrapper
var mockContentRecord = content.SimpleContent{
	Id:            1017738,
	Duration:      10,
	AgeRestricted: false,
	AuthorId:      1,
	CategoryId:    null.Int{},
	Hashtags:      nil,
}

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
	db = database.GetDb()

	kafkaPublishedEvents = nil

	pollTime = 100 * time.Millisecond

	publisherMock = KafkaEventPublisherMock{}

	notifier = comment.NewNotifier(
		pollTime,
		context.TODO(),
		&publisherMock,
		db,
		false,
	)

	contentWrapperMock = &content.ContentWrapperMock{
		GetInternalFn: func(contentIds []int64, includeDeleted bool, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[int64]content.SimpleContent] {
			ch := make(chan wrappers.GenericResponseChan[map[int64]content.SimpleContent], 2)
			go func() {
				defer func() {
					close(ch)
				}()

				if contentIds[0] == mockContentRecord.Id {
					ch <- wrappers.GenericResponseChan[map[int64]content.SimpleContent]{
						Error: nil,
						Response: map[int64]content.SimpleContent{
							mockContentRecord.Id: mockContentRecord,
						},
					}
				}
			}()

			return ch
		},
	}

	os.Exit(m.Run())
}

func baseSetup(t *testing.T) {
	cfg := configs.GetConfig()

	if err := boilerplate_testing.FlushPostgresTables(cfg.Db,
		[]string{"public.comment", "public.comment_vote", "public.content", "public.comment_vote", "public.profile"}, nil, nil); err != nil {
		t.Fatal(err)
	}

	if err := utils.PollutePostgresDatabase(db, "../comments/test_data/seed.json"); err != nil {
		t.Fatal(err)
	}
}

func TestVoteComment(t *testing.T) {
	baseSetup(t)
	vote, err := VoteComment(db, 9700, null.BoolFrom(true), 1, nil, nil, nil, &content.ContentWrapperMock{})
	if err != nil {
		t.Fatal(err)
	}

	a := assert.New(t)
	a.Equal(int64(9700), vote.CommentId)
	a.Equal(int64(1), vote.UserId)
	a.Equal(true, vote.VoteUp.ValueOrZero())

	var comment database.Comment
	if err := db.Where("id = ?", 9700).First(&comment).Error; err != nil {
		t.Fatal(err)
	}
	a.Equal(int64(1), comment.NumUpvotes)
	a.Equal(int64(0), comment.NumDownvotes)

	_, err = VoteComment(db, 9700, null.BoolFrom(true), 1, nil, nil, nil, &content.ContentWrapperMock{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Where("id = ?", 9700).First(&comment).Error; err != nil {
		t.Fatal(err)
	}
	a.Equal(int64(1), comment.NumUpvotes)
	a.Equal(int64(0), comment.NumDownvotes)

	_, err = VoteComment(db, 9700, null.BoolFrom(false), 1, nil, nil, nil, &content.ContentWrapperMock{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Where("id = ?", 9700).First(&comment).Error; err != nil {
		t.Fatal(err)
	}
	a.Equal(int64(0), comment.NumUpvotes)
	a.Equal(int64(1), comment.NumDownvotes)

	_, err = VoteComment(db, 9700, null.NewBool(false, false), 1, nil, nil, nil, &content.ContentWrapperMock{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Where("id = ?", 9700).First(&comment).Error; err != nil {
		t.Fatal(err)
	}
	a.Equal(int64(0), comment.NumUpvotes)
	a.Equal(int64(0), comment.NumDownvotes)

	data, err := VoteComment(db, 9700, null.NewBool(false, false), 1, nil, nil, nil, &content.ContentWrapperMock{})
	a.Nil(data)
	a.Nil(err)

	_, err = VoteComment(db, 9700, null.BoolFrom(false), 1, nil, nil, nil, &content.ContentWrapperMock{})
	if err != nil {
		t.Fatal(err)
	}

	if err := db.Where("id = ?", 9700).First(&comment).Error; err != nil {
		t.Fatal(err)
	}
	a.Equal(int64(0), comment.NumUpvotes)
	a.Equal(int64(1), comment.NumDownvotes)

	_, err = VoteComment(db, 9714, null.BoolFrom(true), 1, nil, nil, nil, &content.ContentWrapperMock{})
	a.NotNil(err)
	a.True(strings.Contains(err.Error(), "record not found"))

	if err := db.Where("id = ?", 9700).First(&comment).Error; err != nil {
		t.Fatal(err)
	}
	a.Equal(int64(0), comment.NumUpvotes)
	a.Equal(int64(1), comment.NumDownvotes)

	_, err = VoteComment(db, 9700, null.BoolFrom(true), 1, notifier, nil, nil, contentWrapperMock)
	if err != nil {
		t.Fatal(err)
	}

	errs := notifier.Flush()
	a.Equal(len(errs), 0)
	a.Greater(len(kafkaPublishedEvents), 0)
}
