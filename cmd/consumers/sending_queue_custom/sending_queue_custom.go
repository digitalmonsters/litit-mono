package sending_queue_custom

import (
	"context"
	"github.com/digitalmonsters/go-common/wrappers/notification_handler"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/segmentio/kafka-go"
)

func process(event newCustomSendingEvent, ctx context.Context, notifySender sender.ISender) (*kafka.Message, error) {
	_, err := notifySender.SendCustomTemplateToUser(notification_handler.NotificationChannelPush,
		event.UserId, "custom", "popup", event.Title, event.Body, event.Headline, nil, ctx)

	if err != nil {
		return nil, err
	}

	return &event.Messages, nil
}
