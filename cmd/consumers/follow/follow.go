package follow

import (
	"context"
	"encoding/json"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/wrappers/notification_handler"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/notification"
	"github.com/digitalmonsters/notification-handler/pkg/renderer"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/gocql/gocql"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
	"time"
)

func process(event newSendingEvent, ctx context.Context, notifySender sender.ISender, userGoWrapper user_go.IUserGoWrapper,
	apmTransaction *apm.Transaction) (*kafka.Message, error) {
	if !event.Follow {
		return &event.Messages, nil
	}

	apm_helper.AddApmLabel(apmTransaction, "user_id", event.UserId)
	apm_helper.AddApmLabel(apmTransaction, "to_user_id", event.ToUserId)

	var userData user_go.UserRecord
	var err error
	var title string
	var body string
	var headline string

	resp := <-userGoWrapper.GetUsers([]int64{event.UserId}, ctx, false)
	if resp.Error != nil {
		return nil, resp.Error.ToError()
	}

	var ok bool
	if userData, ok = resp.Response[event.UserId]; !ok {
		return &event.Messages, errors.WithStack(errors.New("user not found")) // we should continue, no need to retry
	}

	db := database.GetDb(database.DbTypeMaster).WithContext(ctx)

	firstName, lastName := userData.GetFirstAndLastNameWithPrivacy()

	renderingVariables := map[string]string{
		"firstname": firstName,
		"lastname":  lastName,
	}

	renderingVariablesMarshalled, err := json.Marshal(renderingVariables)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var template database.RenderTemplate
	title, body, headline, template, err = notifySender.RenderTemplate(db, "follow", renderingVariables, userData.Language)

	if err == renderer.TemplateRenderingError {
		return &event.Messages, err // we should continue, no need to retry
	} else if err != nil {
		return nil, err
	}

	customData := database.CustomData{"image_url": template.ImageUrl, "route": template.Route, "user_id": event.UserId}
	if _, err = notifySender.SendCustomTemplateToUser(notification_handler.NotificationChannelPush, event.ToUserId, "follow", "user_follow",
		title, body, headline, customData, ctx); err != nil {
		return nil, err
	}

	nt := &database.Notification{
		UserId:        event.ToUserId,
		Type:          "push.profile.following",
		Title:         title,
		Message:       body,
		RelatedUserId: null.IntFrom(event.UserId),
		CreatedAt:     time.Now().UTC(),
		ContentId:     null.IntFrom(0),
		CustomData:    customData,
	}

	if err = db.Create(nt).Error; err != nil {
		return nil, err
	}

	apm_helper.AddApmLabel(apmTransaction, "notification_id", nt.Id.String())

	if err = notification.IncrementUnreadNotificationsCounter(db, event.UserId); err != nil {
		return nil, err
	}

	session := database.GetScyllaSession()

	batch := session.NewBatch(gocql.UnloggedBatch)

	notificationsCount := int64(1)

	if template.IsGrouped {
		notificationRelationIter := session.Query("select user_id from notification_relation where user_id = ? and event_type = ?",
			event.ToUserId, template.Id).WithContext(ctx).Iter()

		var userId int64
		for notificationRelationIter.Scan(&userId) {
			notificationsCount++
		}

		batch.Query("update notification_relation set event_applied = true where user_id = ? and event_type = ? "+
			"and entity_id = ? and related_entity_id = 0", event.ToUserId, template.Id, event.UserId)

		notificationIter := session.Query("select * from notification where user_id = ? and event_type = ? and created_at >= ? limit 1",
			event.ToUserId, template.Id, time.Now().UTC().Add(-3*24*30*time.Hour)).WithContext(ctx).Iter()

		userId = 0
		var eventType string
		var entityId int64
		var relatedEntityId int64
		var createdAt time.Time
		var notificationsCountFromSelect int64

		notificationIter.Scan(&userId, &eventType, &entityId, &relatedEntityId, &createdAt, &notificationsCountFromSelect)

		if err = notificationIter.Close(); err != nil {
			return nil, errors.WithStack(err)
		}

		if notificationsCountFromSelect > notificationsCount {
			notificationsCount = notificationsCountFromSelect + 1
		}
	}

	batch.Query("update notification set notifications_count = ?, title = ?, body = ?, headline = ?, rendering_variables = ?, "+
		"custom_data = ?, image_url = ?, route = ? where user_id = ? and event_type = ? "+
		"and created_at = ? and entity_id = ? and related_entity_id = 0", notificationsCount, title, body, headline,
		string(renderingVariablesMarshalled), string(customDataMarshalled), template.ImageUrl, template.Route,
		event.ToUserId, template.Id, time.Now().UTC(), event.UserId)

	if err := session.ExecuteBatch(batch); err != nil {
		return nil, errors.WithStack(err)
	}

	return &event.Messages, nil
}
