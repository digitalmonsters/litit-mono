package main

import (
	"github.com/digitalmonsters/ads-manager/configs"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/segmentio/kafka-go"
)

var (
	// TODO: read from config
	topicsToVeCreated = []string{"local.views.view_content"}
)

func main() {
	boilerplate.SetupZeroLog()

	cfg := configs.GetConfig()
	hosts := cfg.Wrappers.NotificationHandler.PushPublisher.Hosts
	if hosts == "" {
		panic("kafka hosts not found")
	}
	conn, err := kafka.Dial("tcp", hosts)
	if err != nil {
		panic(err.Error())
	}
	defer conn.Close()

	for _, topic := range topicsToVeCreated {
		err = conn.CreateTopics(kafka.TopicConfig{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		})
		if err != nil {
			panic(err.Error())
		}
	}
}
