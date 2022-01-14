package vote

import (
	"context"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/eventsourcing"
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
	autoFlush bool
}

func NewNotifier(pollTime time.Duration, ctx context.Context,
	eventPublisher eventsourcing.IEventPublisher, autoFlush bool) *Notifier {
	n := &Notifier{
		queueMap:  make(map[string]eventData),
		publisher: eventPublisher,
		mutex:     sync.Mutex{},
		poolTime:  pollTime,
		ctx:       ctx,
		autoFlush: autoFlush,
	}
	if autoFlush{
		n.initQueueListener()
	}
	return n
}

func (s *Notifier) Enqueue(commentId int64, userId int64, voteUp null.Bool, parentId null.Int, commentAuthorId int64,
	comment string, contentId null.Int, profileId null.Int) {
	s.mutex.Lock()

	data := eventData{
		CommentId:       commentId,
		UserId:          userId,
		Upvote:          voteUp,
		ParentId:        parentId,
		CommentAuthorId: commentAuthorId,
		Comment:         comment,
		ContentId:       contentId,
		ProfileId:       profileId,
	}

	s.queueMap[data.GetPublishKey()] = data

	s.mutex.Unlock()
}

func (s *Notifier) Flush() []error {
	apmTransaction := apm_helper.StartNewApmTransaction("votes flush", "publisher", nil, nil)

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

	records := make([]eventsourcing.IEventData, len(queueCopy))

	indexer := 0

	for _, r := range queueCopy {
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
