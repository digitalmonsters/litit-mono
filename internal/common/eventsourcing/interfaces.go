package eventsourcing

import "go.elastic.co/apm"

type IEventData interface {
	GetPublishKey() string
}

type IEventPublisher interface {
	Publish(apmTransaction *apm.Transaction, events ...IEventData) []error
	GetPublisherType() PublisherType
	GetHosts() string
	GetTopic() string
}

type PublisherType int

const (
	PublisherTypeScylla PublisherType = 1
	PublisherTypeKafka  PublisherType = 2
)
