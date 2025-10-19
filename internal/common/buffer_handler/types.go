package common

import (
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/router"
)

type IEventHandler[E eventsourcing.IEventData] interface {
	Process(messages []E, executionData router.MethodExecutionData) []error
	Close() []error
}

type IBufferEventHandler[E eventsourcing.IEventData] interface {
	Enqueue(data ...E)
	Flush() []error
	Close() []error
}
