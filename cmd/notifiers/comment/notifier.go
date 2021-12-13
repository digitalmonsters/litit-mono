package comment

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/comments/pkg/database"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"gopkg.in/guregu/null.v4"
	"sync"
	"time"
)

type Notifier struct {
	queueMap  map[string]eventData
	mutex     sync.Mutex
	publisher eventsourcing.IEventPublisher
	poolTime  time.Duration
	ctx       context.Context
}

func NewNotifier(pollTime time.Duration, ctx context.Context,
	eventPublisher eventsourcing.IEventPublisher) *Notifier {
	n := &Notifier{
		queueMap:  make(map[string]eventData),
		publisher: eventPublisher,
		mutex:     sync.Mutex{},
		poolTime:  pollTime,
		ctx:       ctx,
	}

	n.initQueueListener()

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

	s.queueMap[fmt.Sprintf("%v", comment.Id)] = data

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

	records := make([]eventsourcing.IEventData, len(s.queueMap))
	indexer := 0
	for _, r := range s.queueMap {
		records[indexer] = r

		indexer += 1
	}

	s.queueMap = make(map[string]eventData)

	s.mutex.Unlock()

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
