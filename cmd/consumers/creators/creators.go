package creators

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/wrappers/notification_handler"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/notification"
	"github.com/digitalmonsters/notification-handler/pkg/renderer"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
	"time"
)

func process(event newSendingEvent, ctx context.Context, notifySender sender.ISender, userGoWrapper user_go.IUserGoWrapper) (*kafka.Message, error) {
	var err error

	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "user_id", event.UserId)
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "creator_id", event.Id)

	renderingData := map[string]string{
		"status": fmt.Sprint(event.Status),
	}

	db := database.GetDb(database.DbTypeMaster).WithContext(ctx)

	var title string
	var body string
	var headline string
	var templateName string

	switch event.Status {
	case user_go.CreatorStatusRejected:
		templateName = "creator_status_rejected"
	case user_go.CreatorStatusApproved:
		templateName = "creator_status_approved"
	case user_go.CreatorStatusPending:
		templateName = "creator_status_pending"
	default:
		return &event.Messages, nil
	}

	resp := <-userGoWrapper.GetUsers([]int64{event.UserId}, ctx, false)
	if resp.Error != nil {
		return nil, resp.Error.ToError()
	}

	var userData user_go.UserRecord
	var ok bool

	if userData, ok = resp.Response[event.UserId]; !ok {
		return &event.Messages, errors.WithStack(errors.New("user not found")) // we should continue, no need to retry
	}

	var template database.RenderTemplate
	title, body, headline, template, err = notifySender.RenderTemplate(db, templateName, renderingData, userData.Language)
	if err == renderer.TemplateRenderingError {
		return &event.Messages, err // we should continue, no need to retry
	} else if err != nil {
		return nil, err
	}

	customData := database.CustomData{"image_url": template.ImageUrl, "route": template.Route}

	if _, err = notifySender.SendCustomTemplateToUser(notification_handler.NotificationChannelPush, event.UserId, templateName, "content_creator",
		title, body, headline, customData, ctx); err != nil {
		return nil, err
	}

	if err = db.Create(&database.Notification{
		UserId:               event.UserId,
		Type:                 "push.content-creator.status",
		Title:                title,
		Message:              body,
		CreatedAt:            time.Now().UTC(),
		ContentCreatorStatus: &event.Status,
		RenderingVariables:   renderingData,
		CustomData:           customData,
	}).Error; err != nil {
		return nil, err
	}

	if err = notification.IncrementUnreadNotificationsCounter(db, event.UserId); err != nil {
		return nil, err
	}

	return &event.Messages, nil
}
