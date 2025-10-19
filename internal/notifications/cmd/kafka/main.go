package main

import (
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/notification-handler/configs"
	"github.com/segmentio/kafka-go"
)

var (
	// TODO: read from config
	topicsToVeCreated = []string{
		"local.users.creators",
		"local.notifications.handler_sending_queue",
		"local.comments.comments",
		"local.comments.comment_votes",
		"local.likes.likes",
		"local.content.content",
		"local.users.user",
		"local.notifications.notification_created",
		"local.notifications.push_admin_messages",
		"local.music.creators",
		"local.user.user",
		"local.follows.follows",
	}
)

func main() {
	boilerplate.SetupZeroLog()

	cfg := configs.GetConfig()
	hosts := cfg.Wrappers.NotificationGateway.PushPublisher.Hosts
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
