package kafka_listener

type IKafkaListener interface {
	Close() error
	Listen()
	ListenAsync()
	GetTopic() string
}
