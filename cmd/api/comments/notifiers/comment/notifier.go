package comment

import (
	"context"
	"github.com/digitalmonsters/comments/pkg/database"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"gorm.io/gorm"
	"strconv"
	"sync"
	"time"
)

type Notifier struct {
	queueMap  []eventsourcing.IEventData
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
		queueMap:  make([]eventsourcing.IEventData, 0),
		publisher: eventPublisher,
		mutex:     sync.Mutex{},
		poolTime:  pollTime,
		ctx:       ctx,
		db:        db,
		autoFlush: autoFlush,
	}
	if autoFlush {
		n.initQueueListener()
	}
	return n
}

func (s *Notifier) Enqueue(comment database.Comment, crudOperation eventsourcing.ChangeEvenType,
	crudOperationReason eventsourcing.CommentChangeReason) {
	s.mutex.Lock()

	data := eventsourcing.Comment{
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
		BaseChangeEvent: eventsourcing.BaseChangeEvent{
			CrudOperation:       crudOperation,
			CrudOperationReason: strconv.Itoa(int(crudOperationReason)),
		},
	}

	s.queueMap = append(s.queueMap, data)

	s.mutex.Unlock()
}

func (s *Notifier) Flush() []error {
	apmTransaction := apm_helper.StartNewApmTransaction("content comments flush", "publisher", nil, nil)

	s.mutex.Lock()

	if len(s.queueMap) == 0 {
		apmTransaction.Discard()
		s.mutex.Unlock()
		return nil
	}

	defer apmTransaction.End()

	queueCopy := s.queueMap
	s.queueMap = make([]eventsourcing.IEventData, 0)
	s.mutex.Unlock()

	return s.publisher.Publish(apmTransaction, queueCopy...)
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
