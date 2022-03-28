package notifier

import (
	"context"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/music/pkg/database"
	"github.com/digitalmonsters/music/pkg/global"
	"github.com/rs/zerolog/log"
	"gopkg.in/guregu/null.v4"
	"sync"
	"time"
)

type Notifier struct {
	queue               map[int64]database.Creator
	mutex               *sync.Mutex
	pollTime            time.Duration
	kafkaEventPublisher eventsourcing.IEventPublisher
	ctx                 context.Context
}

func NewService(pollTime time.Duration, eventPublisher eventsourcing.IEventPublisher, ctx context.Context) global.INotifier {
	var mutex sync.Mutex
	n := &Notifier{
		mutex:               &mutex,
		queue:               make(map[int64]database.Creator),
		kafkaEventPublisher: eventPublisher,
		pollTime:            pollTime,
		ctx:                 ctx,
	}

	if boilerplate.GetCurrentEnvironment() != boilerplate.Ci {
		n.initQueueListener()
	}

	return n
}

func (n *Notifier) Enqueue(userId int64, data *database.Creator) {
	if data == nil {
		return
	}

	if data.UserId == 0 {
		return
	}

	n.mutex.Lock()
	n.queue[userId] = *data
	n.mutex.Unlock()
}

func (n *Notifier) Flush() []error {
	apmTransaction := apm_helper.StartNewApmTransaction("music_creator flush", "publisher", nil, nil)
	if len(n.queue) == 0 {
		apmTransaction.Discard()
		return nil
	}

	defer apmTransaction.End()

	n.mutex.Lock()
	queueCopy := n.queue
	n.queue = make(map[int64]database.Creator)
	n.mutex.Unlock()

	records, appErrors := n.handleQueue(queueCopy)

	if len(appErrors) > 0 {
		for _, err := range appErrors {
			apm_helper.CaptureApmError(err, apmTransaction)
		}

		n.updateQueue(queueCopy)

		return appErrors
	}

	if len(records) != 0 {
		publisherErrors := n.kafkaEventPublisher.Publish(apmTransaction, records...)

		if len(publisherErrors) > 0 {
			for _, err := range publisherErrors {
				apm_helper.CaptureApmError(err, apmTransaction)
			}

			n.updateQueue(queueCopy)

			return publisherErrors
		}
	}

	return nil
}

func (n *Notifier) handleQueue(queue map[int64]database.Creator) ([]eventsourcing.IEventData, []error) {
	data := make([]eventsourcing.IEventData, 0)

	for _, event := range queue {
		creatorRequest := eventsourcing.MusicCreatorModel{
			Id:         event.Id,
			UserId:     event.UserId,
			Status:     event.Status,
			LibraryUrl: event.LibraryUrl,
			CreatedAt:  event.CreatedAt,
			ApprovedAt: event.ApprovedAt,
			DeletedAt:  event.DeletedAt,
		}

		if event.Reason != nil {
			creatorRequest.RejectReason = null.StringFrom(event.Reason.Reason)
		}

		data = append(data, creatorRequest)
	}

	return data, nil
}

func (n *Notifier) updateQueue(queue map[int64]database.Creator) {
	n.mutex.Lock()
	for key, value := range queue {
		if _, ok := n.queue[key]; ok {
			continue // we already have that in queue
		} else {
			n.queue[key] = value
		}
	}
	n.mutex.Unlock()
}

func (n *Notifier) Close() []error {
	return n.Flush()
}

func (n *Notifier) initQueueListener() {
	go func() {
		for n.ctx.Err() == nil {
			errs := n.Flush()

			for _, err := range errs {
				log.Error().Str("notifier", "music_creator").Err(err).Send()
			}
			time.Sleep(n.pollTime)
		}
	}()
}
