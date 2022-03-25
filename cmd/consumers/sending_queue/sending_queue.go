package sending_queue

import (
	"context"
	"github.com/digitalmonsters/go-common/wrappers/notification_handler"
	"github.com/digitalmonsters/notification-handler/pkg/renderer"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/segmentio/kafka-go"
)

func process(event newSendingEvent, ctx context.Context, notifySender sender.ISender) (*kafka.Message, error) {
	_, err := notifySender.SendTemplateToUser(notification_handler.NotificationChannelPush,
		event.TemplateName, event.UserId, event.RenderingVariables, ctx)

	if err == renderer.TemplateRenderingError {
		return &event.Messages, err // we should continue, no need to retry
	}

	if err != nil {
		return nil, err
	}

	return &event.Messages, nil
}
