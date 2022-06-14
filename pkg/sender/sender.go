package sender

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/translation"
	"github.com/digitalmonsters/go-common/wrappers/notification_gateway"
	"github.com/digitalmonsters/go-common/wrappers/notification_handler"
	"github.com/digitalmonsters/notification-handler/configs"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/database/scylla"
	notificationPkg "github.com/digitalmonsters/notification-handler/pkg/notification"
	"github.com/digitalmonsters/notification-handler/pkg/renderer"
	"github.com/digitalmonsters/notification-handler/pkg/settings"
	"github.com/digitalmonsters/notification-handler/pkg/token"
	"github.com/gocql/gocql"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"go.elastic.co/apm"
	"gorm.io/gorm"
	"math"
	"strings"
	"time"
)

type Sender struct {
	gateway         notification_gateway.INotificationGatewayWrapper
	settingsService settings.IService
}

func NewSender(gateway notification_gateway.INotificationGatewayWrapper, settingsService settings.IService) *Sender {
	return &Sender{
		gateway:         gateway,
		settingsService: settingsService,
	}
}

func (s *Sender) SendTemplateToUser(channel notification_handler.NotificationChannel,
	title, body, headline string, renderingTemplate database.RenderTemplate, userId int64, renderingData map[string]string,
	customData database.CustomData, ctx context.Context) (interface{}, error) {
	db := database.GetDbWithContext(database.DbTypeReadonly, ctx)

	return s.sendPushTemplateMessageToUser(title, body, headline, renderingTemplate, userId, renderingData, customData, db, ctx)
}

func (s *Sender) SendCustomTemplateToUser(channel notification_handler.NotificationChannel, userId int64, pushType, kind,
	title, body, headline string, customData database.CustomData, isGrouped bool, entityId int64, createdAt time.Time,
	ctx context.Context) (interface{}, error) {
	return s.sendCustomPushTemplateMessageToUser(pushType, kind, title, body, headline, userId, customData, isGrouped,
		entityId, createdAt, ctx)
}

func (s *Sender) SendEmail(msg []notification_gateway.SendEmailMessageRequest, ctx context.Context) error {
	return <-s.gateway.EnqueueEmail(msg, ctx)
}

// TODO: mb repeat logic from sendCustomPushTemplateMessageToUser? or unite methods
func (s *Sender) sendPushTemplateMessageToUser(title, body, headline string,
	renderingTemplate database.RenderTemplate, userId int64, renderingData map[string]string,
	customData map[string]interface{}, db *gorm.DB, ctx context.Context) (interface{}, error) {
	userTokens, err := token.GetUserTokens(db, userId)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	if len(userTokens) == 0 {
		return nil, nil
	}

	isMuted, err := s.settingsService.IsPushNotificationMuted(userId, renderingTemplate.Id, ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if isMuted {
		return nil, nil
	}

	sendResult := <-s.gateway.EnqueuePushForUser(s.preparePushEvents(userTokens, title, body,
		headline, renderingTemplate, fmt.Sprint(userId), renderingData, customData), ctx)

	return nil, sendResult
}

func FloorToNearest(x, floorTo int) int {
	xF := float64(x) / float64(100)
	floorToF := float64(floorTo) / float64(100)
	return int(math.Floor(xF/floorToF) * floorToF * 100)
}

func CeilToNearest(x, floorTo int) int {
	xF := float64(x) / float64(100)
	floorToF := float64(floorTo) / float64(100)
	return int(math.Ceil(xF/floorToF) * floorToF * 100)
}

func (s *Sender) sendGroupedPush(pushType, kind, title, body, headline string, userId int64, entityId int64,
	customData database.CustomData, ctx context.Context) error {
	userTokens, err := token.GetUserTokens(database.GetDbWithContext(database.DbTypeReadonly, ctx), userId)
	if err != nil {
		return errors.WithStack(err)
	}

	if len(userTokens) == 0 {
		return nil
	}

	isMuted, err := s.settingsService.IsPushNotificationMuted(userId, pushType, ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	if isMuted {
		return nil
	}

	session := database.GetScyllaSession()

	notificationRelationIter := session.Query("select user_id, event_applied from notification_relation where user_id = ? "+
		"and event_type = ? and entity_id = ? and related_entity_id = ?", userId, pushType, entityId, 0).Iter()

	var userIdFromSelect int64
	var eventApplied bool
	notificationRelationIter.Scan(&userIdFromSelect, &eventApplied)

	if err = notificationRelationIter.Close(); err != nil {
		return errors.WithStack(err)
	}

	if userIdFromSelect > 0 && !eventApplied {
		return nil
	}

	if err = <-s.gateway.EnqueuePushForUser(s.prepareCustomPushEvents(userTokens, pushType, kind, title, body, headline, fmt.Sprint(userId), customData), ctx); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *Sender) sendCustomPushTemplateMessageToUser(pushType, kind, title, body, headline string, userId int64,
	customData database.CustomData, isGrouped bool, entityId int64, createdAt time.Time, ctx context.Context) (interface{}, error) {
	userTokens, err := token.GetUserTokens(database.GetDbWithContext(database.DbTypeReadonly, ctx), userId)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	if len(userTokens) == 0 {
		return nil, nil
	}

	isMuted, err := s.settingsService.IsPushNotificationMuted(userId, pushType, ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if isMuted {
		return nil, nil
	}

	if !isGrouped {
		sendResult := <-s.gateway.EnqueuePushForUser(s.prepareCustomPushEvents(userTokens, pushType, kind, title, body, headline, fmt.Sprint(userId), customData), ctx)
		return nil, sendResult
	}

	session := database.GetScyllaSession()

	notificationRelationIter := session.Query("select event_applied from notification_relation where user_id = ? "+
		"and event_type = ? and entity_id = ? and related_entity_id = ?", userId, pushType, entityId, 0).Iter()

	var eventApplied bool
	notificationRelationIter.Scan(&eventApplied)

	if err = notificationRelationIter.Close(); err != nil {
		return nil, errors.WithStack(err)
	}

	if !eventApplied {
		return nil, nil
	}

	deadlineKeysLen := (configs.PushNotificationDeadlineKeyMinutes / configs.PushNotificationDeadlineMinutes) * 2
	deadlineKeys := make([]string, deadlineKeysLen)
	newTime := createdAt
	newCurrentMinute := 0

	if newTime.Minute() > configs.PushNotificationDeadlineKeyMinutes {
		newCurrentMinute = configs.PushNotificationDeadlineKeyMinutes
	}

	newTime = time.Date(newTime.Year(), newTime.Month(), newTime.Day(), newTime.Hour(), newCurrentMinute, 0, 0, newTime.Location())
	for i := 0; i < deadlineKeysLen; i++ {
		deadlineKeys[i] = newTime.String()

		if i != deadlineKeysLen-1 {
			newTime = newTime.Add(configs.PushNotificationDeadlineMinutes * time.Minute)
		}
	}

	deadline := createdAt
	minutesDiff := deadline.Minute() - FloorToNearest(deadline.Minute(), 5)
	deadline = deadline.Add(-time.Duration(minutesDiff+configs.PushNotificationDeadlineMinutes*2) * time.Minute)
	deadlines := []string{deadline.String(), deadline.Add(configs.PushNotificationDeadlineMinutes * time.Minute).String(),
		deadline.Add(2 * configs.PushNotificationDeadlineMinutes * time.Minute).String()}

	pushNotificationGroupQueueIter := session.Query(fmt.Sprintf("select deadline_key, deadline, user_id, "+
		"event_type, entity_id, created_at, notification_count from push_notification_group_queue "+
		"where deadline_key in (%v) and deadline in (%v) and user_id = ? and event_type = ? and entity_id = ?",
		strings.Join(deadlineKeys, ","), strings.Join(deadlines, ",")), userId, pushType, entityId).WithContext(ctx).Iter()

	pushNotificationsGroupQueue := make([]scylla.PushNotificationGroupQueue, 0)
	var pushNotificationGroupQueue scylla.PushNotificationGroupQueue
	for pushNotificationGroupQueueIter.Scan(&pushNotificationGroupQueue.DeadlineKey, &pushNotificationGroupQueue.Deadline,
		&pushNotificationGroupQueue.UserId, &pushNotificationGroupQueue.EventType, &pushNotificationGroupQueue.EntityId,
		&pushNotificationGroupQueue.CreatedAt, &pushNotificationGroupQueue.NotificationCount) {
		pushNotificationsGroupQueue = append(pushNotificationsGroupQueue, pushNotificationGroupQueue)
		pushNotificationGroupQueue = scylla.PushNotificationGroupQueue{}
	}

	if err = pushNotificationGroupQueueIter.Close(); err != nil {
		return nil, errors.WithStack(err)
	}

	flooredCreatedAt := time.Date(createdAt.Year(), createdAt.Month(), createdAt.Day(), createdAt.Hour(), FloorToNearest(createdAt.Minute(), 5), 0, 0, createdAt.Location())
	for _, item := range pushNotificationsGroupQueue {
		if !flooredCreatedAt.After(item.CreatedAt.Add(time.Duration(configs.PushNotificationDeadlineKeyMinutes)*time.Minute)) || !item.Deadline.Equal(deadline) {
			continue
		}

		pushNotificationGroupQueue = item
		break
	}

	batch := session.NewBatch(gocql.UnloggedBatch)

	if pushNotificationGroupQueue.UserId == 0 { // empty
		deadline = time.Date(createdAt.Year(), createdAt.Month(), createdAt.Day(), createdAt.Hour(),
			FloorToNearest(createdAt.Minute(), configs.PushNotificationDeadlineMinutes), 0, 0, createdAt.Location())
		batch.Query("update push_notification_group_queue set created_at = ?, notification_count = ? "+
			"where deadline_key = ? and deadline = ? and user_id = ? and event_type = ? and entity_id = ?",
			createdAt, 1, flooredCreatedAt, deadline, userId, pushType, entityId)

		if err = session.ExecuteBatch(batch); err != nil {
			return nil, errors.WithStack(err)
		}

		return nil, nil
	}

	if pushNotificationGroupQueue.Deadline.After(createdAt) {
		return nil, nil
	}

	ceilDeadline := time.Date(pushNotificationGroupQueue.Deadline.Year(), pushNotificationGroupQueue.Deadline.Month(),
		pushNotificationGroupQueue.Deadline.Day(), pushNotificationGroupQueue.Deadline.Hour(),
		CeilToNearest(pushNotificationGroupQueue.Deadline.Minute(), configs.PushNotificationDeadlineKeyMinutes),
		0, 0, pushNotificationGroupQueue.Deadline.Location())
	ceilCurrent := time.Date(createdAt.Year(), createdAt.Month(), createdAt.Day(), createdAt.Hour(),
		CeilToNearest(createdAt.Minute(), configs.PushNotificationDeadlineKeyMinutes), 0, 0, createdAt.Location())

	if !ceilCurrent.After(ceilDeadline) || ceilCurrent.Unix()-ceilDeadline.Unix() > configs.PushNotificationDeadlineMinutes*60 {
		return nil, nil
	}

	batch.Query("update push_notification_group_queue set created_at = ?, notification_count = ? "+
		"where deadline_key = ? and deadline = ? and user_id = ? and event_type = ? and entity_id = ?",
		pushNotificationGroupQueue.CreatedAt, pushNotificationGroupQueue.NotificationCount+1,
		pushNotificationGroupQueue.DeadlineKey, pushNotificationGroupQueue.Deadline, pushNotificationGroupQueue.UserId,
		pushNotificationGroupQueue.EventType, pushNotificationGroupQueue.EntityId)

	if err = session.ExecuteBatch(batch); err != nil {
		return nil, errors.WithStack(err)
	}

	return nil, nil
}

func (s *Sender) preparePushEvents(tokens []database.Device, title string, body string, headline string, template database.RenderTemplate,
	key string, renderingData map[string]string, customData database.CustomData) []notification_gateway.SendPushRequest {
	mm := map[common.DeviceType]*notification_gateway.SendPushRequest{}

	extraData := map[string]string{
		"type":     template.Id,
		"kind":     template.Kind,
		"headline": headline,
	}
	if customData != nil {
		js, err := json.Marshal(&customData)
		if err != nil {
			log.Error().Str("push_type", template.Id).Str("push_kind", template.Kind).Err(err).Send()
		}
		extraData["custom_data"] = string(js)
	}

	for k, v := range renderingData {
		if _, ok := extraData[k]; !ok {
			extraData[k] = v
		}
	}

	for _, t := range tokens {
		if _, ok := mm[t.Platform]; !ok {
			req := notification_gateway.SendPushRequest{
				Tokens:     nil,
				DeviceType: t.Platform,
				Title:      title,
				Body:       body,
				ExtraData:  extraData,
				PublishKey: key,
			}

			mm[t.Platform] = &req
		}

		mm[t.Platform].Tokens = append(mm[t.Platform].Tokens, t.PushToken)
	}

	var resp []notification_gateway.SendPushRequest

	for _, v := range mm {
		resp = append(resp, *v)
	}

	return resp
}

func (s *Sender) prepareCustomPushEvents(tokens []database.Device, pushType, kind, title string, body string, headline string,
	key string, customData database.CustomData) []notification_gateway.SendPushRequest {
	mm := map[common.DeviceType]*notification_gateway.SendPushRequest{}

	var extraData = map[string]string{
		"type":     pushType,
		"kind":     kind,
		"headline": headline,
	}
	if customData != nil {
		js, err := json.Marshal(&customData)
		if err != nil {
			log.Error().Str("push_type", pushType).Str("push_kind", kind).Err(err).Send()
		}
		extraData["custom_data"] = string(js)
	}

	for _, t := range tokens {
		if _, ok := mm[t.Platform]; !ok {
			req := notification_gateway.SendPushRequest{
				Tokens:     nil,
				DeviceType: t.Platform,
				Title:      title,
				Body:       body,
				ExtraData:  extraData,
				PublishKey: key,
			}

			mm[t.Platform] = &req
		}

		mm[t.Platform].Tokens = append(mm[t.Platform].Tokens, t.PushToken)
	}

	var resp []notification_gateway.SendPushRequest

	for _, v := range mm {
		resp = append(resp, *v)
	}

	return resp
}

func (s *Sender) RenderTemplate(db *gorm.DB, templateName string, renderingData map[string]string,
	language translation.Language) (title string, body string, headline string, renderingTemplate database.RenderTemplate, err error) {
	var renderTemplate database.RenderTemplate

	if err := db.Where("id = ?", strings.ToLower(templateName)).Take(&renderTemplate).Error; err != nil {
		return "", "", "", renderTemplate, errors.WithStack(err)
	}

	title, body, headline, err = renderer.Render(renderTemplate, renderingData, language)

	return title, body, headline, renderTemplate, err
}

func (s *Sender) PushNotification(notification database.Notification, entityId int64, relatedEntityId int64,
	templateName string, language translation.Language, customKind string, ctx context.Context) (shouldRetry bool, innerErr error) {
	var template database.RenderTemplate
	var title string
	var body string
	var headline string
	var err error

	title, body, headline, template, err = s.RenderTemplate(database.GetDb(database.DbTypeMaster).WithContext(ctx),
		templateName, notification.RenderingVariables, language)

	if err == renderer.TemplateRenderingError {
		return false, errors.WithStack(err) // we should continue, no need to retry
	} else if err != nil {
		return true, errors.WithStack(err)
	}

	notification.CreatedAt = time.Now().UTC()

	notification.CustomData["image_url"] = template.ImageUrl
	notification.CustomData["route"] = template.Route

	customDataMarshalled, err := json.Marshal(notification.CustomData)
	if err != nil {
		return false, errors.WithStack(err)
	}

	var notificationInfoMarshalled []byte
	notificationInfoMarshalled, err = json.Marshal(notification)
	if err != nil {
		return false, errors.WithStack(err)
	}

	var renderingVariablesMarshalled []byte
	renderingVariablesMarshalled, err = json.Marshal(notification.RenderingVariables)
	if err != nil {
		return false, errors.WithStack(err)
	}

	kind := ""
	if len(customKind) == 0 {
		kind = template.Kind
	} else {
		kind = customKind
	}

	session := database.GetScyllaSession()

	batch := session.NewBatch(gocql.UnloggedBatch)

	notificationsCount := int64(1)

	if template.IsGrouped {
		notificationRelationIter := session.Query("select user_id from notification_relation where user_id = ? and event_type = ?",
			notification.UserId, template.Id).WithContext(ctx).Iter()

		var userIdSelected int64
		for notificationRelationIter.Scan(&userIdSelected) {
			notificationsCount++
		}

		batch.Query("update notification_relation set event_applied = true where user_id = ? and event_type = ? "+
			"and entity_id = ? and related_entity_id = ?", notification.UserId, template.Id, entityId, relatedEntityId)

		notificationIter := session.Query("select user_id, event_type, entity_id, related_entity_id, created_at, "+
			"notifications_count from notification where user_id = ? and event_type = ? and created_at >= ? limit 1",
			notification.UserId, template.Id, notification.CreatedAt.Add(-3*24*30*time.Hour)).WithContext(ctx).Iter()

		userIdSelected = 0
		var eventType string
		var entityIdSelected int64
		var relatedEntityIdSelected int64
		var createdAt time.Time
		var notificationsCountSelected int64

		notificationIter.Scan(&userIdSelected, &eventType, &entityIdSelected, &relatedEntityIdSelected, &createdAt, &notificationsCountSelected)

		if err = notificationIter.Close(); err != nil {
			return true, errors.WithStack(err)
		}

		if notificationsCountSelected > notificationsCount {
			notificationsCount = notificationsCountSelected + 1
			// TODO: rerender text somehow for notificationsCount > 1
		}
	}

	notification.Title = title
	notification.Message = body
	notification.NotificationsCount = notificationsCount

	batch.Query("update notification set notifications_count = ?, title = ?, body = ?, headline = ?, kind = ?, rendering_variables = ?, "+
		"custom_data = ?, notification_info = ? where user_id = ? and event_type = ? "+
		"and created_at = ? and entity_id = ? and related_entity_id = ?", notificationsCount, title, body, headline,
		kind, string(renderingVariablesMarshalled), string(customDataMarshalled), string(notificationInfoMarshalled),
		notification.UserId, template.Id, notification.CreatedAt, entityId, relatedEntityId)

	if err = session.ExecuteBatch(batch); err != nil {
		return true, errors.WithStack(err)
	}

	tx := database.GetDb(database.DbTypeMaster).WithContext(ctx).Begin()
	defer tx.Rollback()

	if err = tx.Create(&notification).Error; err != nil {
		return true, err
	}

	if err = notificationPkg.IncrementUnreadNotificationsCounter(tx, notification.UserId); err != nil {
		return true, err
	}

	if err = tx.Commit().Error; err != nil {
		return true, errors.WithStack(err)
	}

	if _, err = s.SendCustomTemplateToUser(notification_handler.NotificationChannelPush, notification.UserId,
		template.Id, kind, title, body, headline, notification.CustomData, template.IsGrouped, entityId,
		notification.CreatedAt, ctx); err != nil {
		return true, errors.WithStack(err)
	}

	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "notification_id", notification.Id.String())

	return false, nil
}

// TODO: create job to trigger this every PushNotificationDeadlineMinutes
func (s *Sender) checkPushNotificationDeadlineMinutes(ctx context.Context) error {
	session := database.GetScyllaSession()

	createdAt := time.Now().UTC()
	deadlineKeysLen := (configs.PushNotificationDeadlineKeyMinutes / configs.PushNotificationDeadlineMinutes) * 2
	deadlineKeys := make([]string, deadlineKeysLen)
	newTime := createdAt
	newCurrentMinute := 0

	if newTime.Minute() > configs.PushNotificationDeadlineKeyMinutes {
		newCurrentMinute = configs.PushNotificationDeadlineKeyMinutes
	}

	newTime = time.Date(newTime.Year(), newTime.Month(), newTime.Day(), newTime.Hour(), newCurrentMinute, 0, 0, newTime.Location())
	for i := 0; i < deadlineKeysLen; i++ {
		deadlineKeys[i] = newTime.String()

		if i != deadlineKeysLen-1 {
			newTime = newTime.Add(configs.PushNotificationDeadlineMinutes * time.Minute)
		}
	}

	deadline := createdAt
	minutesDiff := deadline.Minute() - FloorToNearest(deadline.Minute(), 5)
	deadline = deadline.Add(-time.Duration(minutesDiff+configs.PushNotificationDeadlineMinutes*2) * time.Minute)
	deadlines := []string{deadline.String(), deadline.Add(configs.PushNotificationDeadlineMinutes * time.Minute).String(),
		deadline.Add(2 * configs.PushNotificationDeadlineMinutes * time.Minute).String()}

	pushNotificationGroupQueueIter := session.Query(fmt.Sprintf("select deadline_key, deadline, user_id, "+
		"event_type, entity_id, created_at, notification_count from push_notification_group_queue "+
		"where deadline_key in (%v) and deadline in (%v)",
		strings.Join(deadlineKeys, ","), strings.Join(deadlines, ","))).WithContext(ctx).Iter()

	pushNotificationsGroupQueue := make([]scylla.PushNotificationGroupQueue, 0)
	var pushNotificationGroupQueue scylla.PushNotificationGroupQueue
	for pushNotificationGroupQueueIter.Scan(&pushNotificationGroupQueue.DeadlineKey, &pushNotificationGroupQueue.Deadline,
		&pushNotificationGroupQueue.UserId, &pushNotificationGroupQueue.EventType, &pushNotificationGroupQueue.EntityId,
		&pushNotificationGroupQueue.CreatedAt, &pushNotificationGroupQueue.NotificationCount) {
		pushNotificationsGroupQueue = append(pushNotificationsGroupQueue, pushNotificationGroupQueue)
		pushNotificationGroupQueue = scylla.PushNotificationGroupQueue{}
	}

	if err := pushNotificationGroupQueueIter.Close(); err != nil {
		return errors.WithStack(err)
	}

	flooredCreatedAt := time.Date(createdAt.Year(), createdAt.Month(), createdAt.Day(), createdAt.Hour(), FloorToNearest(createdAt.Minute(), 5), 0, 0, createdAt.Location())

	for _, item := range pushNotificationsGroupQueue {
		// TODO: move cycle body to separate func
		// TODO: pass new ctx and use new apmTransaction
		if flooredCreatedAt.After(item.CreatedAt.Add(time.Duration(configs.PushNotificationDeadlineKeyMinutes)*time.Minute)) &&
			item.CreatedAt.Add(time.Duration(configs.PushNotificationDeadlineKeyMinutes+configs.PushNotificationDeadlineMinutes)*time.Minute).After(flooredCreatedAt) {
			notificationIter := session.Query("select user_id, related_entity_id, title, body, headline, kind, rendering_variables, custom_data "+
				"from notification where user_id = ? and event_type = ? and created_at = ? and entity_id = ? limit 1",
				item.UserId, item.EventType, item.CreatedAt, item.EntityId).WithContext(ctx).Iter()

			notification := scylla.Notification{
				UserId:    item.UserId,
				EventType: item.EventType,
				EntityId:  item.EntityId,
				CreatedAt: item.CreatedAt,
			}

			var userIdFromSelect int64
			notificationIter.Scan(userIdFromSelect, &notification.RelatedEntityId, &notification.Title, &notification.Body,
				&notification.Headline, &notification.Kind, &notification.RenderingVariables, &notification.CustomData)

			if err := notificationIter.Close(); err != nil {
				apm_helper.LogError(errors.WithStack(err), ctx)
				continue
			}

			if userIdFromSelect == 0 { // should not happen
				continue
			}

			var customData database.CustomData
			if err := json.Unmarshal([]byte(notification.CustomData), &customData); err != nil {
				return errors.WithStack(err)
			}

			if err := s.sendGroupedPush(item.EventType, notification.Kind, notification.Title, notification.Body, notification.Headline,
				item.UserId, item.EntityId, customData, ctx); err != nil {
				return errors.WithStack(err)
			}

			batch := session.NewBatch(gocql.UnloggedBatch)

			batch.Query("delete from push_notification_group_queue where deadline_key = ? and deadline = ? "+
				"and user_id = ? and event_type = ? and entity_id = ?", item.DeadlineKey, item.Deadline, item.UserId,
				item.EventType, item.EntityId)

			if err := session.ExecuteBatch(batch); err != nil {
				apm_helper.LogError(errors.WithStack(err), ctx)
				continue
			}
		}

		ceilDeadline := time.Date(pushNotificationGroupQueue.Deadline.Year(), pushNotificationGroupQueue.Deadline.Month(),
			pushNotificationGroupQueue.Deadline.Day(), pushNotificationGroupQueue.Deadline.Hour(),
			CeilToNearest(pushNotificationGroupQueue.Deadline.Minute(), configs.PushNotificationDeadlineKeyMinutes),
			0, 0, pushNotificationGroupQueue.Deadline.Location())
		ceilCurrent := time.Date(createdAt.Year(), createdAt.Month(), createdAt.Day(), createdAt.Hour(),
			CeilToNearest(createdAt.Minute(), configs.PushNotificationDeadlineKeyMinutes), 0, 0, createdAt.Location())

		if !ceilCurrent.After(ceilDeadline) || ceilCurrent.Unix()-ceilDeadline.Unix() > configs.PushNotificationDeadlineMinutes*60 {
			return nil
		}

		// TODO: refactor remove duplicate code
		notificationIter := session.Query("select user_id, related_entity_id, title, body, headline, kind, rendering_variables, custom_data "+
			"from notification where user_id = ? and event_type = ? and created_at = ? and entity_id = ? limit 1",
			item.UserId, item.EventType, item.CreatedAt, item.EntityId).WithContext(ctx).Iter()

		notification := scylla.Notification{
			UserId:    item.UserId,
			EventType: item.EventType,
			EntityId:  item.EntityId,
			CreatedAt: item.CreatedAt,
		}

		var userIdFromSelect int64
		notificationIter.Scan(userIdFromSelect, &notification.RelatedEntityId, &notification.Title, &notification.Body,
			&notification.Headline, &notification.Kind, &notification.RenderingVariables, &notification.CustomData)

		if err := notificationIter.Close(); err != nil {
			apm_helper.LogError(errors.WithStack(err), ctx)
			continue
		}

		if userIdFromSelect == 0 { // should not happen
			continue
		}

		var customData database.CustomData
		if err := json.Unmarshal([]byte(notification.CustomData), &customData); err != nil {
			return errors.WithStack(err)
		}

		if err := s.sendGroupedPush(item.EventType, notification.Kind, notification.Title, notification.Body, notification.Headline,
			item.UserId, item.EntityId, customData, ctx); err != nil {
			return errors.WithStack(err)
		}

		batch := session.NewBatch(gocql.UnloggedBatch)

		batch.Query("delete from push_notification_group_queue where deadline_key = ? and deadline = ? "+
			"and user_id = ? and event_type = ? and entity_id = ?", item.DeadlineKey, item.Deadline, item.UserId,
			item.EventType, item.EntityId)

		if err := session.ExecuteBatch(batch); err != nil {
			apm_helper.LogError(errors.WithStack(err), ctx)
			continue
		}
	}

	return nil
}
