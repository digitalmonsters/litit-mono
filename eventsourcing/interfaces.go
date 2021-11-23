package eventsourcing

import "go.elastic.co/apm"

type IEventData interface {
	GetPublishKey() string
}

type IEventPublisher interface {
	Publish(apmTransaction *apm.Transaction, events ...IEventData) []error
	GetPublisherType() PublisherType
}

type PublisherType int

const (
	PublisherTypeScylla PublisherType = 1
	PublisherTypeKafka  PublisherType = 2
)
