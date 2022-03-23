package music_creators

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/notification-handler/pkg/renderer"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/segmentio/kafka-go"
)

func process(event newSendingEvent, ctx context.Context, notifySender sender.ISender) (*kafka.Message, error) {
	var err error

	renderingData := map[string]string{
		"status": fmt.Sprint(event.Status),
	}

	switch event.Status {
	case eventsourcing.CreatorStatusRejected:
		_, err = notifySender.SendTemplateToUser(
			sender.NotificationChannelPush,
			"music_creator_status_rejected",
			event.UserId,
			renderingData,
			ctx,
		)
	case eventsourcing.CreatorStatusApproved:
		_, err = notifySender.SendTemplateToUser(
			sender.NotificationChannelPush,
			"music_creator_status_approved",
			event.UserId,
			renderingData,
			ctx,
		)
	case eventsourcing.CreatorStatusPending:
		_, err = notifySender.SendTemplateToUser(
			sender.NotificationChannelPush,
			"music_creator_status_pending",
			event.UserId,
			renderingData,
			ctx,
		)
	default:
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
