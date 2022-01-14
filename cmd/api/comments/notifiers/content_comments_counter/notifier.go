package content_comments_counter

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/eventsourcing"
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

func (s *Notifier) Enqueue(contentId int64, contentCommentsCount int64) {
	s.mutex.Lock()

	s.queueMap[fmt.Sprintf("%v", contentId)] = eventData{
		ContentId: contentId,
		Count:     contentCommentsCount,
	}

	s.mutex.Unlock()
}

func (s *Notifier) Flush() []error {
	apmTransaction := apm_helper.StartNewApmTransaction("content comments counter flush", "publisher", nil, nil)
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
