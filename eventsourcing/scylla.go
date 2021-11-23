package eventsourcing

import (
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/gammazero/workerpool"
	"github.com/gocql/gocql"
	"github.com/pkg/errors"
	"go.elastic.co/apm"
	"sync"
)

type ScyllaEventPublisher struct {
	cluster       *gocql.ClusterConfig
	session       *gocql.Session
	sesMutex      sync.Mutex
	workerPoll    *workerpool.WorkerPool
	batchSize     int
	publisherType PublisherType
}

func NewScyllaEventPublisher(configuration boilerplate.ScyllaConfiguration, batchSize int, workerPoolSize int) *ScyllaEventPublisher {
	return &ScyllaEventPublisher{
		cluster:       boilerplate.GetScyllaCluster(configuration),
		batchSize:     batchSize,
		workerPoll:    workerpool.New(workerPoolSize),
		publisherType: PublisherTypeScylla,
	}
}

func (s *ScyllaEventPublisher) getSession() (*gocql.Session, error) {
	if s.session != nil {
		return s.session, nil
	}

	s.sesMutex.Lock()
	defer s.sesMutex.Unlock()

	if s.session != nil {
		return s.session, nil
	}

	if session, err := s.cluster.CreateSession(); err != nil {
		return nil, errors.WithStack(err)
	} else {
		s.session = session
	}

	return s.session, nil
}

type ScylaQuery string

func (e ScylaQuery) GetPublishKey() string {
	return ""
}

func (s *ScyllaEventPublisher) Publish(apmTransaction *apm.Transaction, events ...IEventData) []error {
	session, err := s.getSession()

	if err != nil {
		return []error{errors.WithStack(err)}
	}

	batch := session.NewBatch(gocql.UnloggedBatch)
	batchChannels := make([]chan error, 0)

	var span *apm.Span
	if apmTransaction != nil {
		span = apmTransaction.StartSpan("new scylla event publishing", "scylla", nil)
	}

	internalErrors := make([]error, 0)

	for _, event := range events {
		q, ok := event.(ScylaQuery)
		if !ok {
			er := errors.New("can't convert event to query")
			apm_helper.CaptureApmError(er, apmTransaction)
			internalErrors = append(internalErrors, er)
			continue
		}

		batch.Query(string(q))

		if len(batch.Entries) > s.batchSize {
			b := batch

			ch := make(chan error, 2)
			batchChannels = append(batchChannels, ch)

			s.workerPoll.Submit(func() {
				if err := session.ExecuteBatch(b); err != nil {
					ch <- err
				}
				close(ch)
			})

			batch = session.NewBatch(gocql.UnloggedBatch)
		}
	}

	if len(batch.Entries) > 0 {
		ch := make(chan error, 2)
		batchChannels = append(batchChannels, ch)

		s.workerPoll.Submit(func() {
			if err := session.ExecuteBatch(batch); err != nil {
				ch <- err
			}
			close(ch)
		})
	}

	for _, c := range batchChannels {
		if err := <-c; err != nil {
			internalErrors = append(internalErrors, err)
		}

		if len(internalErrors) > 0 {
			return internalErrors
		}
	}

	if span != nil {
		span.Context.SetLabel("count", len(batchChannels))
		span.End()
	}

	return internalErrors
}

func (s *ScyllaEventPublisher) GetPublisherType() PublisherType {
	return s.publisherType
}
