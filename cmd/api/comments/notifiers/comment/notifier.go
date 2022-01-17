package comment

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/comments/pkg/database"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/thoas/go-funk"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"sync"
	"time"
)

type Notifier struct {
	queueMap  map[string]eventData
	mutex     sync.Mutex
	publisher eventsourcing.IEventPublisher
	poolTime  time.Duration
	ctx       context.Context
	db        *gorm.DB
	autoFlush bool
}

func NewNotifier(pollTime time.Duration, ctx context.Context,
	eventPublisher eventsourcing.IEventPublisher, db *gorm.DB, autoFlush bool) *Notifier {
	n := &Notifier{
		queueMap:  make(map[string]eventData),
		publisher: eventPublisher,
		mutex:     sync.Mutex{},
		poolTime:  pollTime,
		ctx:       ctx,
		db:        db,
		autoFlush: autoFlush,
	}
	if autoFlush{
		n.initQueueListener()
	}
	return n
}

func (s *Notifier) Enqueue(comment database.Comment, content content.SimpleContent, eventType EventType) {
	s.mutex.Lock()

	data := eventData{
		Id:           comment.Id,
		AuthorId:     comment.AuthorId,
		NumReplies:   comment.NumReplies,
		NumUpvotes:   comment.NumUpvotes,
		NumDownvotes: comment.NumDownvotes,
		CreatedAt:    comment.CreatedAt,
		Active:       comment.Active,
		Comment:      comment.Comment,
		ContentId:    comment.ContentId,
		ParentId:     comment.ParentId,
		ProfileId:    comment.ProfileId,
		EventType:    eventType,
	}

	if eventType == ContentResourceTypeCreate || eventType == ContentResourceTypeUpdate || eventType == ContentResourceTypeDelete {
		data.Width = null.IntFrom(int64(content.Width))
		data.Height = null.IntFrom(int64(content.Height))
		data.VideoId = null.StringFrom(content.VideoId)
	}

	data.ContentAuthorId = null.IntFrom(content.AuthorId)

	s.queueMap[fmt.Sprintf("%v", comment.Id)] = data

	s.mutex.Unlock()
}

func (s *Notifier) Flush() []error {
	apmTransaction := apm_helper.StartNewApmTransaction("content comments flush", "publisher", nil, nil)
	ctx := apm.ContextWithTransaction(context.TODO(), apmTransaction)

	s.mutex.Lock()

	if len(s.queueMap) == 0 {
		apmTransaction.Discard()
		s.mutex.Unlock()
		return nil
	}

	defer apmTransaction.End()

	queueCopy := s.queueMap
	s.queueMap = make(map[string]eventData)
	s.mutex.Unlock()

	var parentCommentIds []int64
	for _, v := range queueCopy {
		parentId := v.ParentId.ValueOrZero()

		if parentId > 0 {
			if !funk.ContainsInt64(parentCommentIds, parentId) {
				parentCommentIds = append(parentCommentIds, parentId)
			}
		}
	}

	parentComments := map[int64]database.Comment{}

	if len(parentCommentIds) > 0 {
		var tempComments []database.Comment

		if err := s.db.WithContext(ctx).Where("id in ?", parentCommentIds).Find(&tempComments).Error; err != nil {
			apm_helper.CaptureApmError(err, apmTransaction)
		} else {
			for _, v := range tempComments {
				parentComments[v.Id] = v
			}
		}
	}

	records := make([]eventsourcing.IEventData, len(queueCopy))

	indexer := 0

	for _, r := range queueCopy {
		parentId := r.ParentId.ValueOrZero()

		if v, ok := parentComments[parentId]; ok {
			r.ParentAuthorId = null.IntFrom(v.AuthorId)
		}

		records[indexer] = r

		indexer += 1
	}
	return s.publisher.Publish(apmTransaction, records...)
}

func (s *Notifier) Close() []error {
	return s.Flush()
}

func (s *Notifier) initQueueListener() {
	go func() {
		for s.ctx.Err() == nil {
			_ = s.Flush()
			time.Sleep(s.poolTime)
		}
	}()
}
