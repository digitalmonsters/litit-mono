package sending_queue_custom

import (
	"context"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/segmentio/kafka-go"
)

func process(event newCustomSendingEvent, ctx context.Context, notifySender sender.ISender) (*kafka.Message, error) {
	_, err := notifySender.SendCustomTemplateToUser(sender.NotificationChannelPush, event.UserId, event.Title, event.Body, event.Headline, ctx)

	if err != nil {
		return nil, err
	}

	return &event.Messages, nil
}
