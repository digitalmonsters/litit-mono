package sender

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/translation"
	"github.com/digitalmonsters/go-common/wrappers/notification_gateway"
	"github.com/digitalmonsters/notification-handler/configs"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/database/scylla"
	notificationPkg "github.com/digitalmonsters/notification-handler/pkg/notification"
	"github.com/digitalmonsters/notification-handler/pkg/renderer"
	"github.com/digitalmonsters/notification-handler/pkg/settings"
	"github.com/digitalmonsters/notification-handler/pkg/token"
	"github.com/digitalmonsters/notification-handler/pkg/utils"
	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"go.elastic.co/apm"
	"go.elastic.co/apm/module/apmhttp"
	"math"
	"strconv"
	"strings"
	"time"
)

type Sender struct {
	gateway         notification_gateway.INotificationGatewayWrapper
	settingsService settings.IService
	jobber          *machinery.Server
}

func NewSender(gateway notification_gateway.INotificationGatewayWrapper, settingsService settings.IService, jobber *machinery.Server) *Sender {
	return &Sender{
		gateway:         gateway,
		settingsService: settingsService,
		jobber:          jobber,
	}
}

func (s *Sender) SendEmail(msg []notification_gateway.SendEmailMessageRequest, ctx context.Context) error {
	return <-s.gateway.EnqueueEmail(msg, ctx)
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
	deadlineKeys := make([]time.Time, deadlineKeysLen)
	newTime := time.Date(createdAt.Year(), createdAt.Month(), createdAt.Day(), createdAt.Hour(),
		FloorToNearest(createdAt.Minute(), configs.PushNotificationDeadlineKeyMinutes), 0, 0, createdAt.Location())

	for i := 0; i < deadlineKeysLen; i++ {
		deadlineKeys[i] = newTime

		if i != deadlineKeysLen-1 {
			newTime = newTime.Add(configs.PushNotificationDeadlineMinutes * time.Minute)
		}
	}

	deadline := createdAt
	minutesDiff := deadline.Minute() - FloorToNearest(deadline.Minute(), configs.PushNotificationDeadlineMinutes)
	deadline = deadline.Add(-time.Duration(minutesDiff+configs.PushNotificationDeadlineMinutes*2) * time.Minute)
	deadline = time.Date(deadline.Year(), deadline.Month(), deadline.Day(), deadline.Hour(), deadline.Minute(), 0, 0, deadline.Location())
	deadlines := []time.Time{deadline, deadline.Add(configs.PushNotificationDeadlineMinutes * time.Minute),
		deadline.Add(2 * configs.PushNotificationDeadlineMinutes * time.Minute)}

	pushNotificationGroupQueueIter := session.Query(fmt.Sprintf("select deadline_key, deadline, user_id, "+
		"event_type, entity_id, created_at, notification_count from push_notification_group_queue "+
		"where deadline_key in (%v) and deadline in (%v) and user_id = ? and event_type = ? and entity_id = ?",
		utils.JoinDatesForInStatement(deadlineKeys), utils.JoinDatesForInStatement(deadlines)), userId, pushType, entityId).WithContext(ctx).Iter()

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

	flooredCreatedAt := time.Date(createdAt.Year(), createdAt.Month(), createdAt.Day(), createdAt.Hour(),
		FloorToNearest(createdAt.Minute(), configs.PushNotificationDeadlineMinutes), 0, 0, createdAt.Location())
	for _, item := range pushNotificationsGroupQueue {
		if !flooredCreatedAt.After(item.CreatedAt.Add(time.Duration(configs.PushNotificationDeadlineKeyMinutes)*time.Minute)) || !item.Deadline.Equal(deadline) {
			continue
		}

		pushNotificationGroupQueue = item
		break
	}

	batch := session.NewBatch(gocql.UnloggedBatch).WithContext(ctx)

	if pushNotificationGroupQueue.UserId == 0 { // empty
		deadline = time.Date(createdAt.Year(), createdAt.Month(), createdAt.Day(), createdAt.Hour(),
			FloorToNearest(createdAt.Minute(), configs.PushNotificationDeadlineMinutes)+configs.PushNotificationDeadlineMinutes, 0, 0, createdAt.Location())
		deadlineKey := time.Date(createdAt.Year(), createdAt.Month(), createdAt.Day(), createdAt.Hour(),
			CeilToNearest(createdAt.Minute(), configs.PushNotificationDeadlineKeyMinutes), 0, 0, createdAt.Location())
		batch.Query("update push_notification_group_queue set created_at = ?, notification_count = ? "+
			"where deadline_key = ? and deadline = ? and user_id = ? and event_type = ? and entity_id = ?",
			createdAt, 1, deadlineKey, deadline, userId, pushType, entityId)

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
		pushNotificationGroupQueue.DeadlineKey, pushNotificationGroupQueue.Deadline.Add(configs.PushNotificationDeadlineMinutes*time.Minute),
		pushNotificationGroupQueue.UserId, pushNotificationGroupQueue.EventType, pushNotificationGroupQueue.EntityId)

	if err = session.ExecuteBatch(batch); err != nil {
		return nil, errors.WithStack(err)
	}

	return nil, nil
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

func (s *Sender) PushNotification(notification database.Notification, entityId int64, relatedEntityId int64,
	templateName string, language translation.Language, customKind string, ctx context.Context) (shouldRetry bool, innerErr error) {
	var template database.RenderTemplate
	var title string
	var body string
	var err error
	isCustomPush := strings.Contains(templateName, "push_admin")

	if !isCustomPush {
		db := database.GetDbWithContext(database.DbTypeMaster, ctx)

		if err = db.Where("id = ?", templateName).Find(&template).Error; err != nil {
			return true, errors.WithStack(err)
		}

		if template.Id != templateName {
			return false, errors.WithStack(errors.New("template not found"))
		}
	} else {
		title = notification.Title
		body = notification.Message
	}

	template.Id = templateName

	notification.Id = uuid.New()
	notification.CreatedAt = time.Now().UTC()

	if notification.CustomData == nil {
		notification.CustomData = make(database.CustomData)
	}

	if !isCustomPush {
		if len(template.ImageUrl) > 0 {
			notification.CustomData["image_url"] = template.ImageUrl
		}

		if len(template.Route) > 0 {
			notification.CustomData["route"] = template.Route
		}
	}

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

	batch := session.NewBatch(gocql.UnloggedBatch).WithContext(ctx)

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

		notificationIter := session.Query("select user_id, entity_id, related_entity_id, created_at, "+
			"notifications_count from notification where user_id = ? and event_type = ? and created_at >= ? limit 1",
			notification.UserId, template.Id, notification.CreatedAt.Add(-3*24*30*time.Hour)).WithContext(ctx).Iter()

		userIdSelected = 0
		var entityIdSelected int64
		var relatedEntityIdSelected int64
		var createdAt time.Time
		var notificationsCountSelected int64

		notificationIter.Scan(&userIdSelected, &entityIdSelected, &relatedEntityIdSelected, &createdAt, &notificationsCountSelected)

		if err = notificationIter.Close(); err != nil {
			return true, errors.WithStack(err)
		}

		if userIdSelected != 0 {
			batch.Query("delete from notification where user_id = ? and event_type = ? and created_at = ? and entity_id = ? and related_entity_id = ?",
				notification.UserId, template.Id, createdAt, entityIdSelected, relatedEntityIdSelected)
		}

		if notificationsCountSelected > notificationsCount {
			notificationsCount = notificationsCountSelected + 1
		}
	}

	var headline string
	var titleMultiple string
	var bodyMultiple string
	var headlineMultiple string

	if notification.RenderingVariables == nil {
		notification.RenderingVariables = database.RenderingVariables{}
	}

	notification.RenderingVariables["notificationsCount"] = strconv.FormatInt(notificationsCount, 10)

	if !isCustomPush {
		title, body, headline, titleMultiple, bodyMultiple, headlineMultiple, err = renderer.Render(template, notification.RenderingVariables, language)
		if err == renderer.TemplateRenderingError {
			return false, errors.WithStack(err) // we should continue, no need to retry
		} else if err != nil {
			return true, errors.WithStack(err)
		}
	}

	notification.Title = title
	notification.Message = body

	if notificationsCount > 1 {
		title = titleMultiple
		body = bodyMultiple
		headline = headlineMultiple
	}

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

	if _, err = s.sendCustomPushTemplateMessageToUser(template.Id, kind, title, body, headline, notification.UserId, notification.CustomData, template.IsGrouped,
		entityId, notification.CreatedAt, ctx); err != nil {
		return true, errors.WithStack(err)
	}

	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "notification_id", notification.Id.String())

	return false, nil
}

func (s *Sender) CheckPushNotificationDeadlineMinutes(ctx context.Context) error {
	session := database.GetScyllaSession()

	currentDate := time.Now().UTC()
	deadlineKeysLen := (configs.PushNotificationDeadlineKeyMinutes / configs.PushNotificationDeadlineMinutes) * 2
	deadlineKeys := make([]time.Time, deadlineKeysLen)
	newTime := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), currentDate.Hour(),
		FloorToNearest(currentDate.Minute(), configs.PushNotificationDeadlineKeyMinutes), 0, 0, currentDate.Location())

	for i := 0; i < deadlineKeysLen; i++ {
		deadlineKeys[i] = newTime

		if i != deadlineKeysLen-1 {
			newTime = newTime.Add(configs.PushNotificationDeadlineMinutes * time.Minute)
		}
	}

	deadline := currentDate
	minutesDiff := deadline.Minute() - FloorToNearest(deadline.Minute(), configs.PushNotificationDeadlineMinutes)
	deadline = deadline.Add(-time.Duration(minutesDiff+configs.PushNotificationDeadlineMinutes*2) * time.Minute)
	deadline = time.Date(deadline.Year(), deadline.Month(), deadline.Day(), deadline.Hour(), deadline.Minute(), 0, 0, deadline.Location())
	deadlines := []time.Time{deadline, deadline.Add(configs.PushNotificationDeadlineMinutes * time.Minute),
		deadline.Add(2 * configs.PushNotificationDeadlineMinutes * time.Minute)}

	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "deadline_key", deadlineKeys)
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "deadline", deadlines)

	pushNotificationGroupQueueIter := session.Query(fmt.Sprintf("select deadline_key, deadline, user_id, "+
		"event_type, entity_id, created_at, notification_count from push_notification_group_queue "+
		"where deadline_key in (%v) and deadline in (%v)",
		utils.JoinDatesForInStatement(deadlineKeys), utils.JoinDatesForInStatement(deadlines))).WithContext(ctx).Iter()

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

	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "grouped_queued_notifications_count", len(pushNotificationsGroupQueue))
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "push_notifications_group_queue", pushNotificationsGroupQueue)

	for _, item := range pushNotificationsGroupQueue {
		itemMarshalled, _ := json.Marshal(item)
		if _, err := s.jobber.SendTask(&tasks.Signature{
			Name: string(configs.UserPushNotificationTask),
			Args: []tasks.Arg{
				{
					Name:  "currentDate",
					Type:  "string",
					Value: currentDate.String(),
				},
				{
					Name:  "item",
					Type:  "string",
					Value: string(itemMarshalled),
				},
				{
					Name:  "traceHeader",
					Type:  "string",
					Value: "",
				},
			}}); err != nil {
			apm_helper.LogError(err, ctx)
			return err
		}
	}

	return nil
}

func (s *Sender) getNotificationForGroupSend(userId int64, eventType string, createdAt time.Time, entityId int64,
	ctx context.Context) (*scylla.Notification, error) {
	session := database.GetScyllaSession()

	notificationIter := session.Query("select user_id, related_entity_id, title, body, headline, kind, rendering_variables, custom_data "+
		"from notification where user_id = ? and event_type = ? and created_at = ? and entity_id = ? limit 1",
		userId, eventType, createdAt, entityId).WithContext(ctx).Iter()

	notification := scylla.Notification{
		UserId:    userId,
		EventType: eventType,
		EntityId:  entityId,
		CreatedAt: createdAt,
	}

	var userIdFromSelect int64
	notificationIter.Scan(userIdFromSelect, &notification.RelatedEntityId, &notification.Title, &notification.Body,
		&notification.Headline, &notification.Kind, &notification.RenderingVariables, &notification.CustomData)

	if err := notificationIter.Close(); err != nil {
		return nil, errors.WithStack(err)
	}

	if userIdFromSelect == 0 { // should not happen
		return nil, errors.WithStack(errors.New("notification not found"))
	}

	return &notification, nil
}

func (s *Sender) deleteNotificationFromQueue(deadlineKey time.Time, deadline time.Time, userId int64, eventType string,
	entityId int64, ctx context.Context) error {
	session := database.GetScyllaSession()
	batch := session.NewBatch(gocql.UnloggedBatch).WithContext(ctx)

	batch.Query("delete from push_notification_group_queue where deadline_key = ? and deadline = ? "+
		"and user_id = ? and event_type = ? and entity_id = ?", deadlineKey, deadline, userId,
		eventType, entityId)

	if err := session.ExecuteBatch(batch); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *Sender) updateNotificationQueueAndSendPush(deadlineKey time.Time, deadline time.Time, userId int64,
	eventType string, createdAt time.Time, entityId int64, ctx context.Context) error {
	notification, err := s.getNotificationForGroupSend(userId, eventType, createdAt, entityId, ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	var customData database.CustomData
	if err := json.Unmarshal([]byte(notification.CustomData), &customData); err != nil {
		return errors.WithStack(err)
	}

	if err := s.sendGroupedPush(eventType, notification.Kind, notification.Title, notification.Body, notification.Headline,
		userId, entityId, customData, ctx); err != nil {
		return errors.WithStack(err)
	}

	if err := s.deleteNotificationFromQueue(deadlineKey, deadline, userId,
		eventType, entityId, ctx); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *Sender) SendDeadlinedNotification(currentDate time.Time, item scylla.PushNotificationGroupQueue, ctx context.Context) (shouldLog bool, innerErr error) {
	flooredCreatedAt := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), currentDate.Hour(),
		FloorToNearest(currentDate.Minute(), configs.PushNotificationDeadlineMinutes), 0, 0, currentDate.Location())

	if flooredCreatedAt.After(item.CreatedAt.Add(time.Duration(configs.PushNotificationDeadlineKeyMinutes)*time.Minute)) &&
		item.CreatedAt.Add(time.Duration(configs.PushNotificationDeadlineKeyMinutes+configs.PushNotificationDeadlineMinutes)*time.Minute).After(flooredCreatedAt) {
		if err := s.updateNotificationQueueAndSendPush(item.DeadlineKey, item.Deadline, item.UserId, item.EventType,
			item.CreatedAt, item.EntityId, ctx); err != nil {
			return true, errors.WithStack(err)
		}

		return true, nil
	}

	ceilDeadline := time.Date(item.Deadline.Year(), item.Deadline.Month(),
		item.Deadline.Day(), item.Deadline.Hour(),
		CeilToNearest(item.Deadline.Minute(), configs.PushNotificationDeadlineMinutes),
		0, 0, item.Deadline.Location())
	ceilCurrent := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), currentDate.Hour(),
		CeilToNearest(currentDate.Minute(), configs.PushNotificationDeadlineMinutes), 0, 0, currentDate.Location())

	if !ceilCurrent.After(ceilDeadline) || ceilCurrent.Unix()-ceilDeadline.Unix() > configs.PushNotificationDeadlineMinutes*60 {
		return false, nil
	}

	if err := s.updateNotificationQueueAndSendPush(item.DeadlineKey, item.Deadline, item.UserId, item.EventType,
		item.CreatedAt, item.EntityId, ctx); err != nil {
		return true, errors.WithStack(err)
	}

	return true, nil
}

func (s *Sender) RegisterUserPushNotificationTasks() error {
	if err := s.jobber.RegisterTask(string(configs.UserPushNotificationTask),
		func(currentDate string, item string, traceHeader string) error {
			var apmTransaction *apm.Transaction

			if parsed, err := apmhttp.ParseTraceparentHeader(traceHeader); err != nil {
				log.Err(err).Send()
				apmTransaction = apm_helper.StartNewApmTransaction(string(configs.UserPushNotificationTask),
					"push_notification", nil, nil)
			} else {
				apmTransaction = apm_helper.StartNewApmTransactionWithTraceData(string(configs.UserPushNotificationTask),
					"push_notification", nil, parsed)
			}

			shouldLog := false
			var err error

			defer func() {
				if shouldLog {
					apmTransaction.End()
				} else {
					apmTransaction.Discard()
				}
			}()

			ctx := boilerplate.CreateCustomContext(context.Background(), apmTransaction, log.Logger)

			var currentDateUnmarshalled time.Time
			currentDateUnmarshalled, err = time.Parse("2006-01-02 15:04:05 -0700 UTC", currentDate)
			if err != nil {
				apm_helper.LogError(errors.WithStack(err), ctx)
				return errors.WithStack(err)
			}

			var itemUnmarshalled scylla.PushNotificationGroupQueue
			if err = json.Unmarshal([]byte(item), &itemUnmarshalled); err != nil {
				apm_helper.LogError(errors.WithStack(err), ctx)
				return errors.WithStack(err)
			}

			apm_helper.AddApmLabel(apmTransaction, "current_date", currentDateUnmarshalled)
			apm_helper.AddApmLabel(apmTransaction, "deadline_key", itemUnmarshalled.DeadlineKey)
			apm_helper.AddApmLabel(apmTransaction, "deadline", itemUnmarshalled.Deadline)
			apm_helper.AddApmLabel(apmTransaction, "user_id", itemUnmarshalled.UserId)
			apm_helper.AddApmLabel(apmTransaction, "event_type", itemUnmarshalled.EventType)
			apm_helper.AddApmLabel(apmTransaction, "entity_id", itemUnmarshalled.EntityId)
			apm_helper.AddApmLabel(apmTransaction, "created_at", itemUnmarshalled.CreatedAt)
			apm_helper.AddApmLabel(apmTransaction, "notification_count", itemUnmarshalled.NotificationCount)

			if shouldLog, err = s.SendDeadlinedNotification(currentDateUnmarshalled, itemUnmarshalled, ctx); err != nil {
				apm_helper.LogError(errors.WithStack(err), ctx)
				return errors.WithStack(err)
			}

			return nil
		}); err != nil {
		return err
	}

	if err := s.jobber.RegisterTask(string(configs.GeneralPushNotificationTask), func() error {
		apmTransaction := apm_helper.StartNewApmTransaction(string(configs.GeneralPushNotificationTask),
			"push_notification", nil, nil)

		defer apmTransaction.End()

		ctx := boilerplate.CreateCustomContext(context.Background(), apmTransaction, log.Logger)

		if err := s.CheckPushNotificationDeadlineMinutes(ctx); err != nil {
			apm_helper.LogError(errors.WithStack(err), ctx)
			return errors.WithStack(err)
		}

		return nil
	}); err != nil {
		return err
	}

	if err := s.jobber.RegisterPeriodicTask(fmt.Sprintf("*/%v * * * *", configs.PushNotificationDeadlineMinutes),
		string(configs.PeriodicPushNotificationTask), &tasks.Signature{
			Name: string(configs.GeneralPushNotificationTask),
		}); err != nil {
		return err
	}

	return nil
}
