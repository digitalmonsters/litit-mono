package common

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/router"
	"github.com/rs/zerolog/log"
	"sync"
	"time"
)

type BufferHandler[T eventsourcing.IEventData] struct {
	buffer        map[string]T
	mx            sync.Mutex
	flushInterval time.Duration
	eventHandler  IEventHandler[T]
	name          string
	appCtx        context.Context
	cancelCtx     context.Context
	cancelFn      context.CancelFunc
	autoFlush     bool
}

func NewBufferHandler[T eventsourcing.IEventData](name string, pollTime time.Duration, eventHandler IEventHandler[T], autoFlush bool, ctx context.Context) *BufferHandler[T] {

	cancelCtx, cancelFn := context.WithCancel(ctx)

	n := &BufferHandler[T]{
		buffer:        make(map[string]T),
		mx:            sync.Mutex{},
		flushInterval: pollTime,
		eventHandler:  eventHandler,
		name:          name,
		autoFlush:     autoFlush,
		appCtx:        ctx,
		cancelFn:      cancelFn,
		cancelCtx:     cancelCtx,
	}
	if autoFlush {
		n.initQueueListener()
	}
	return n
}

func (s *BufferHandler[T]) Enqueue(data ...T) {
	s.mx.Lock()
	for _, ev := range data {
		s.buffer[ev.GetPublishKey()] = ev
	}
	s.mx.Unlock()
}

func (s *BufferHandler[T]) Flush() []error {
	apmTransaction := apm_helper.StartNewApmTransaction(fmt.Sprintf("%s flush", s.name), "buffer handler", nil, nil)
	s.mx.Lock()

	if len(s.buffer) == 0 {
		apmTransaction.Discard()
		s.mx.Unlock()
		return nil
	}

	defer apmTransaction.End()

	records := make([]T, len(s.buffer))
	indexer := 0
	for _, r := range s.buffer {
		records[indexer] = r
		indexer++
	}
	s.buffer = make(map[string]T)
	s.mx.Unlock()

	ctx := boilerplate.CreateCustomContext(s.appCtx, apmTransaction, log.Logger)

	return s.eventHandler.Process(records, router.MethodExecutionData{
		ApmTransaction: apmTransaction,
		Context:        ctx,
	})
}

func (s *BufferHandler[T]) Close() []error {
	s.cancelFn()
	errs := s.Flush()
	notifierErrs := s.eventHandler.Close()
	return append(errs, notifierErrs...)
}

func (s *BufferHandler[T]) initQueueListener() {
	go func() {
		for s.cancelCtx.Err() == nil {
			_ = s.Flush()
			time.Sleep(s.flushInterval)
		}
	}()
}
