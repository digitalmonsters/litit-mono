package tokenomics_notification

import (
	"context"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/wrappers/notification_handler"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	notificationPkg "github.com/digitalmonsters/notification-handler/pkg/notification"
	"github.com/digitalmonsters/notification-handler/pkg/renderer"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
	"time"
)

func process(event newSendingEvent, ctx context.Context, notifySender sender.ISender, userGoWrapper user_go.IUserGoWrapper,
	apmTransaction *apm.Transaction) (*kafka.Message, error) {
	var err error
	var rendererData map[string]string
	var templateName string
	var templateType string
	contentId := null.Int{}

	apm_helper.AddApmLabel(apmTransaction, "event_type", string(event.Type))
	apm_helper.AddApmLabel(apmTransaction, "user_id", event.Payload.UserId)
	apm_helper.AddApmLabel(apmTransaction, "related_user_id", event.Payload.RelatedUserId.ValueOrZero())

	switch event.Type {
	case eventsourcing.TokenomicsNotificationTip:
		var userData user_go.UserRecord

		resp := <-userGoWrapper.GetUsers([]int64{event.Payload.RelatedUserId.ValueOrZero()}, apmTransaction, false)
		if resp.Error != nil {
			return nil, resp.Error.ToError()
		}

		var ok bool
		if userData, ok = resp.Items[event.Payload.RelatedUserId.ValueOrZero()]; !ok {
			return &event.Messages, errors.WithStack(errors.New("user not found")) // we should continue, no need to retry
		}

		templateName = "tip"
		templateType = "push.tip"

		firstName, lastName := userData.GetFirstAndLastNameWithPrivacy()

		rendererData = map[string]string{
			"firstname":    firstName,
			"lastname":     lastName,
			"pointsAmount": event.Payload.PointsAmount.Decimal.String(),
		}

		contentId = null.IntFrom(0)
	case eventsourcing.TokenomicsNotificationDailyBonusTime:
		templateName = "bonus_time"
		templateType = "push.bonus.daily"
	case eventsourcing.TokenomicsNotificationDailyBonusFollowers:
		templateName = "bonus_followers"
		templateType = "push.bonus.followers"
	default:
		return &event.Messages, nil
	}

	db := database.GetDb(database.DbTypeMaster).WithContext(ctx)

	var title string
	var body string
	var headline string

	title, body, headline, _, err = notifySender.RenderTemplate(db, templateName, rendererData)
	if err == renderer.TemplateRenderingError {
		return &event.Messages, err // we should continue, no need to retry
	} else if err != nil {
		return nil, err
	}

	if _, err = notifySender.SendCustomTemplateToUser(notification_handler.NotificationChannelPush, event.Payload.UserId, templateName, "default",
		title, body, headline, nil, ctx); err != nil {
		return nil, err
	}

	notification := database.Notification{
		UserId:             event.Payload.UserId,
		Type:               templateType,
		Title:              title,
		Message:            body,
		CreatedAt:          time.Now().UTC(),
		ContentId:          contentId,
		RelatedUserId:      event.Payload.RelatedUserId,
		RenderingVariables: rendererData,
	}

	if err = db.Create(&notification).Error; err != nil {
		return nil, err
	}

	apm_helper.AddApmLabel(apmTransaction, "notification_id", notification.Id.String())

	if err = notificationPkg.IncrementUnreadNotificationsCounter(db, event.Payload.UserId); err != nil {
		return nil, err
	}

	return &event.Messages, nil
}
