package creators

import (
	"context"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/notification-handler/pkg/renderer"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/segmentio/kafka-go"
)

func process(event newSendingEvent, ctx context.Context, notifySender sender.ISender) (*kafka.Message, error) {
	var err error
	if event.Status == eventsourcing.CreatorStatusRejected {
		_, err = notifySender.SendTemplateToUser(
			sender.NotificationChannelPush,
			"creator_status_rejected",
			event.UserId,
			map[string]string{},
			ctx,
		)
	} else if event.Status == eventsourcing.CreatorStatusApproved {
		_, err = notifySender.SendTemplateToUser(
			sender.NotificationChannelPush,
			"creator_status_approved",
			event.UserId,
			map[string]string{},
			ctx,
		)
	} else {
		return &event.Messages, nil
	}

	if err == renderer.TemplateRenderingError {
		return &event.Messages, err // we should continue, no need to retry
	}
	if err != nil {
		return nil, err
	}

	return &event.Messages, nil
}
