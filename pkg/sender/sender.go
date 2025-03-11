package sender

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/translation"
	"github.com/digitalmonsters/go-common/wrappers/notification_gateway"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/notification-handler/configs"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/database/scylla"
	"github.com/digitalmonsters/notification-handler/pkg/firebase"
	notificationPkg "github.com/digitalmonsters/notification-handler/pkg/notification"
	"github.com/digitalmonsters/notification-handler/pkg/renderer"
	"github.com/digitalmonsters/notification-handler/pkg/settings"
	"github.com/digitalmonsters/notification-handler/pkg/token"
	"github.com/digitalmonsters/notification-handler/pkg/utils"
	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
	"go.elastic.co/apm"
	"go.elastic.co/apm/module/apmhttp"
	"gorm.io/gorm/clause"
)

type Sender struct {
	gateway         notification_gateway.INotificationGatewayWrapper
	settingsService settings.IService
	jobber          *machinery.Server
	userWrapper     user_go.IUserGoWrapper
	firebaseClient  *firebase.FirebaseClient
}

func NewSender(gateway notification_gateway.INotificationGatewayWrapper, settingsService settings.IService,
	jobber *machinery.Server, userWrapper user_go.IUserGoWrapper, firebaseClient *firebase.FirebaseClient) *Sender {
	sender := &Sender{
		gateway:         gateway,
		settingsService: settingsService,
		jobber:          jobber,
		userWrapper:     userWrapper,
		firebaseClient:  firebaseClient,
	}
	setupNotificationCronJobs(sender)
	return sender
}

func setupNotificationCronJobs(sender *Sender) {
	c := cron.New()

	c.AddFunc("0 4 * * *", func() {
		ctx := context.Background()
		log.Info().Msg("Running daily notification aggregation job (8pm PST)")
		if err := sender.SendDailyAggregatedNotifications(ctx); err != nil {
			log.Error().Err(err).Msg("Error running daily notification aggregation job")
		}
	})

	c.Start()
}

// SendDailyAggregatedNotifications is called by a cron job once a day
func (s *Sender) SendDailyAggregatedNotifications(ctx context.Context) error {
	log.Ctx(ctx).Info().Msg("Starting daily notification aggregation job")

	db := database.GetDbWithContext(database.DbTypeMaster, ctx)

	var userIds []int64
	if err := db.Model(&database.Notification{}).
		Where("type IN (?) AND aggregated_sent = false",
			[]string{"push.profile.following", "push.content.like"}).
		Distinct().
		Pluck("user_id", &userIds).Error; err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to get users with notifications")
		return err
	}

	log.Ctx(ctx).Info().Int("user_count", len(userIds)).Msg("Found users with notifications")

	for _, userId := range userIds {
		go func(uid int64) {
			if err := s.processUserNotifications(uid, ctx); err != nil {
				log.Ctx(ctx).Error().Err(err).Int64("user_id", uid).
					Msg("Failed to process notifications")
			}
		}(userId)
	}

	return nil
}

// Process notifications for a single user
func (s *Sender) processUserNotifications(userId int64, ctx context.Context) error {
	db := database.GetDbWithContext(database.DbTypeMaster, ctx)

	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer tx.Rollback()

	var notificationCount int64
	if err := tx.Model(&database.Notification{}).
		Where("user_id = ? AND type IN (?) AND aggregated_sent = false",
			userId, []string{"push.profile.following", "push.content.like"}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Count(&notificationCount).Error; err != nil {
		return err
	}

	if notificationCount == 0 {
		tx.Rollback()
		return nil
	}

	var friendRequestCount int64
	if err := tx.Model(&database.Notification{}).
		Where("user_id = ? AND type = ? AND aggregated_sent = false AND message LIKE '%friend request%'",
			userId, "push.profile.following").
		Count(&friendRequestCount).Error; err != nil {
		return err
	}

	var followCount int64
	if err := tx.Model(&database.Notification{}).
		Where("user_id = ? AND type = ? AND aggregated_sent = false AND message LIKE '%following you%'",
			userId, "push.profile.following").
		Count(&followCount).Error; err != nil {
		return err
	}

	var likesCount int64
	if err := tx.Model(&database.Notification{}).
		Where("user_id = ? AND type = ? AND aggregated_sent = false",
			userId, "push.content.like").
		Count(&likesCount).Error; err != nil {
		return err
	}

	totalCount := friendRequestCount + followCount + likesCount
	if totalCount == 0 {
		tx.Rollback()
		return nil
	}

	var messageBuilder strings.Builder
	var notificationTitle string
	var activityTypes []string

	if friendRequestCount > 0 {
		activityTypes = append(activityTypes, fmt.Sprintf("%d friend request(s)", friendRequestCount))
	}
	if followCount > 0 {
		activityTypes = append(activityTypes, fmt.Sprintf("%d new follower(s)", followCount))
	}
	if likesCount > 0 {
		activityTypes = append(activityTypes, fmt.Sprintf("%d like(s)", likesCount))
	}

	notificationTitle = "Lit.it"

	if len(activityTypes) > 1 {
		messageBuilder.WriteString("You have ")
		for i, activity := range activityTypes {
			if i > 0 {
				if i == len(activityTypes)-1 {
					messageBuilder.WriteString(" and ")
				} else {
					messageBuilder.WriteString(", ")
				}
			}
			messageBuilder.WriteString(activity)
		}
	} else if friendRequestCount > 0 {
		messageBuilder.WriteString(fmt.Sprintf("You have %d friend request(s)", friendRequestCount))
	} else if followCount > 0 {
		messageBuilder.WriteString(fmt.Sprintf("You have %d new follower(s)", followCount))
	} else if likesCount > 0 {
		messageBuilder.WriteString(fmt.Sprintf("You got %d like(s) on your videos", likesCount))
	}

	if err := tx.Model(&database.Notification{}).
		Where("user_id = ? AND type IN (?) AND aggregated_sent = false",
			userId, []string{"push.profile.following", "push.content.like"}).
		Update("aggregated_sent", true).Error; err != nil {
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	deviceInfo, err := notificationPkg.GetLatestDeviceForUser(int(userId), db)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to get device for user")
		return err
	}

	if deviceInfo.PushToken == "" {
		log.Ctx(ctx).Info().Int64("user_id", userId).Msg("User has no push token, skipping")
		return nil
	}

	data := map[string]string{
		"friend_requests_count": fmt.Sprintf("%d", friendRequestCount),
		"follows_count":         fmt.Sprintf("%d", followCount),
		"likes_count":           fmt.Sprintf("%d", likesCount),
		"is_aggregated":         "true",
	}

	// Send notification
	fResp, err := s.firebaseClient.SendNotification(
		ctx,
		deviceInfo.PushToken,
		string(deviceInfo.Platform),
		"",
		notificationTitle,
		"",
		messageBuilder.String(),
		"push.aggregate",
		data,
	)

	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to send Firebase notification")
		return err
	}

	log.Ctx(ctx).Info().
		Str("response", fResp).
		Int64("user_id", userId).
		Msg("Sent aggregated notification successfully")

	return nil
}

func (s *Sender) SendEmail(msg []notification_gateway.SendEmailMessageRequest, ctx context.Context) error {
	return <-s.gateway.EnqueueEmail(msg, ctx)
}

func TimeToNearestMinutes(date time.Time, roundToMinutes int, floorOrCeil bool) time.Time {
	if roundToMinutes <= 0 {
		return time.Date(date.Year(), date.Month(), date.Day(), date.Hour(), date.Minute(), 0, 0, date.Location())
	}

	if roundToMinutes < 60 { // 1h
		minutes := date.Minute()

		if floorOrCeil {
			minutes = FloorToNearest(minutes, roundToMinutes)
		} else {
			minutes = CeilToNearest(minutes, roundToMinutes)
		}

		return time.Date(date.Year(), date.Month(), date.Day(), date.Hour(), minutes, 0, 0, date.Location())
	}

	if roundToMinutes >= 60 && roundToMinutes < 1440 { // 1h and 24h
		hours := date.Hour()

		if floorOrCeil {
			hours = FloorToNearest(hours, roundToMinutes/60)
		} else {
			hours = CeilToNearest(hours, roundToMinutes/60)
		}

		return time.Date(date.Year(), date.Month(), date.Day(), hours, roundToMinutes%60, 0, 0, date.Location())
	}

	if roundToMinutes >= 1440 && roundToMinutes < 43200 { // 24h and 30 days
		days := date.Day()

		if floorOrCeil {
			days = FloorToNearest(days, roundToMinutes/1440) + 1
		} else {
			days = CeilToNearest(days, roundToMinutes/1440) + 1
		}

		return time.Date(date.Year(), date.Month(), days, 0, roundToMinutes%1440, 0, 0, date.Location())
	}

	if roundToMinutes >= 43200 && roundToMinutes < 518400 { // 30 days and 12 months
		months := int(date.Month())

		if floorOrCeil {
			months = FloorToNearest(months, roundToMinutes/43200) + 1
		} else {
			months = CeilToNearest(months, roundToMinutes/43200) + 1
		}

		return time.Date(date.Year(), time.Month(months), 1, 0, roundToMinutes%43200, 0, 0, date.Location())
	}

	years := date.Year()

	if floorOrCeil {
		years = FloorToNearest(years, roundToMinutes/518400)
	} else {
		years = CeilToNearest(years, roundToMinutes/518400)
	}

	return time.Date(years, 1, 1, 0, roundToMinutes%518400, 0, 0, date.Location())
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

func (s *Sender) sendGroupedPush(eventType, kind string, userId int64, entityId int64, notificationCount int64,
	renderingVariables database.RenderingVariables, customData database.CustomData, ctx context.Context) error {
	db := database.GetDbWithContext(database.DbTypeReadonly, ctx)

	var template database.RenderTemplate
	if err := db.Where("id = ?", eventType).Find(&template).Error; err != nil {
		return errors.WithStack(err)
	}

	if template.Id != eventType {
		return errors.WithStack(errors.New("template not found"))
	}

	if template.Muted {
		return nil
	}

	userTokens, err := token.GetUserTokens(database.GetDbWithContext(database.DbTypeReadonly, ctx), userId)
	if err != nil {
		return errors.WithStack(err)
	}

	if len(userTokens) == 0 {
		return nil
	}

	isMuted, err := s.settingsService.IsPushNotificationMuted(userId, eventType, ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	if isMuted {
		return nil
	}

	session := database.GetScyllaSession()

	notificationRelationIter := session.Query("select user_id, event_applied from notification_relation where user_id = ? "+
		"and event_type = ? and entity_id = ? and related_entity_id = ?", userId, eventType, entityId, 0).Iter()

	var userIdFromSelect int64
	var eventApplied bool
	notificationRelationIter.Scan(&userIdFromSelect, &eventApplied)

	if err = notificationRelationIter.Close(); err != nil {
		return errors.WithStack(err)
	}

	if userIdFromSelect > 0 && !eventApplied {
		return nil
	}

	var userData user_go.UserRecord

	resp := <-s.userWrapper.GetUsers([]int64{userId}, ctx, false)
	if resp.Error != nil {
		return errors.WithStack(resp.Error.ToError())
	}

	var ok bool
	if userData, ok = resp.Response[userId]; !ok {
		return errors.WithStack(errors.New("user not found"))
	}

	renderingVariables["notificationsCount"] = strconv.FormatInt(notificationCount-1, 10)

	if firstname, ok := renderingVariables["firstname"]; ok && len(strings.TrimSpace(firstname)) == 0 {
		if notificationCount <= 1 {
			return nil
		} else {
			renderingVariables["firstname"] = "Someone"
		}
	}

	var title string
	var body string
	var headline string
	var titleMultiple string
	var bodyMultiple string
	var headlineMultiple string
	title, body, headline, titleMultiple, bodyMultiple, headlineMultiple, err = renderer.Render(template, renderingVariables, userData.Language)
	if err != nil {
		return errors.WithStack(err)
	}

	if notificationCount > 1 {
		title = titleMultiple
		body = bodyMultiple
		headline = headlineMultiple
	}

	if err = <-s.gateway.EnqueuePushForUser(s.prepareCustomPushEvents(userTokens, eventType, kind, title, body, headline, fmt.Sprint(userId), customData,
		userId), ctx); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *Sender) sendCustomPushTemplateMessageToUser(pushType, kind, title, body, headline string, userId int64,
	customData database.CustomData, isGrouped bool, entityId int64, relatedEntityId int64, createdAt time.Time,
	ctx context.Context) (interface{}, error) {
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
		sendResult := <-s.gateway.EnqueuePushForUser(s.prepareCustomPushEvents(userTokens, pushType, kind, title, body, headline, fmt.Sprint(userId), customData,
			userId), ctx)
		return nil, sendResult
	}

	session := database.GetScyllaSession()

	notificationRelationIter := session.Query("select event_applied from notification_relation where user_id = ? "+
		"and event_type = ? and entity_id = ?", userId, pushType, entityId).Iter()

	var eventApplied bool
	notificationRelationIter.Scan(&eventApplied)

	if err = notificationRelationIter.Close(); err != nil {
		return nil, errors.WithStack(err)
	}

	if !eventApplied {
		return nil, nil
	}

	deadlineKeys, deadlines := GetDeadlinesForSelect(createdAt)
	deadline := deadlines[0]

	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "deadline_key", deadlineKeys)
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "deadline", deadlines)

	query := fmt.Sprintf("select deadline_key, deadline, user_id, "+
		"event_type, entity_id, created_at, notification_count from push_notification_group_queue "+
		"where deadline_key in (%v) and deadline in (%v) and user_id = ? and event_type = ?",
		utils.JoinDatesForInStatement(deadlineKeys), utils.JoinDatesForInStatement(deadlines))

	if relatedEntityId != 0 {
		query = fmt.Sprintf("%v and entity_id = %v", query, entityId)
	}

	pushNotificationGroupQueueIter := session.Query(query, userId, pushType).WithContext(ctx).Iter()

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

	flooredCreatedAt := TimeToNearestMinutes(createdAt, configs.PushNotificationDeadlineMinutes, true)

	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "grouped_queued_notifications_count", len(pushNotificationsGroupQueue))
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "push_notifications_group_queue", pushNotificationsGroupQueue)

	for _, item := range pushNotificationsGroupQueue {
		if !flooredCreatedAt.Before(item.DeadlineKey) /* >= */ || item.Deadline.Equal(deadline) {
			continue
		}

		pushNotificationGroupQueue = item
		break
	}

	batch := session.NewBatch(gocql.UnloggedBatch).WithContext(ctx)

	hasPreviousPushNotificationGroupQueueItem := pushNotificationGroupQueue.UserId != 0
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "has_previous_push_notification_group_queue_item", hasPreviousPushNotificationGroupQueueItem)

	if !hasPreviousPushNotificationGroupQueueItem {
		deadline = TimeToNearestMinutes(createdAt, configs.PushNotificationDeadlineMinutes, true).
			Add(configs.PushNotificationDeadlineMinutes * time.Minute)

		var deadlineKey time.Time
		if configs.PushNotificationDeadlineKeyMinutes != configs.PushNotificationDeadlineMinutes {
			deadlineKey = createdAt.Add(configs.PushNotificationDeadlineKeyMinutes * time.Minute)
			deadlineKey = TimeToNearestMinutes(deadlineKey, configs.PushNotificationDeadlineMinutes, false)
		} else {
			deadlineKey = deadline
		}

		batch.Query("update push_notification_group_queue set created_at = ?, notification_count = ? "+
			"where deadline_key = ? and deadline = ? and user_id = ? and event_type = ? and entity_id = ?",
			createdAt, 1, deadlineKey, deadline, userId, pushType, entityId)

		if err = session.ExecuteBatch(batch); err != nil {
			return nil, errors.WithStack(err)
		}

		return nil, nil
	}

	notificationCount := pushNotificationGroupQueue.NotificationCount + 1

	newDeadline := pushNotificationGroupQueue.Deadline.Add(configs.PushNotificationDeadlineMinutes * time.Minute)

	if notificationCount <= 2 || newDeadline.After(pushNotificationGroupQueue.DeadlineKey) {
		newDeadline = pushNotificationGroupQueue.Deadline
	}

	if relatedEntityId == 0 {
		pushNotificationGroupQueue.EntityId = entityId
		if err = session.Query("delete from push_notification_group_queue where deadline_key = ? and deadline = ? and "+
			"user_id = ? and event_type = ?", pushNotificationGroupQueue.DeadlineKey, pushNotificationGroupQueue.Deadline,
			pushNotificationGroupQueue.UserId, pushNotificationGroupQueue.EventType).Exec(); err != nil {
			return nil, errors.WithStack(err)
		}
	}

	batch.Query("update push_notification_group_queue set created_at = ?, notification_count = ? "+
		"where deadline_key = ? and deadline = ? and user_id = ? and event_type = ? and entity_id = ?",
		pushNotificationGroupQueue.CreatedAt, notificationCount, pushNotificationGroupQueue.DeadlineKey, newDeadline,
		pushNotificationGroupQueue.UserId, pushNotificationGroupQueue.EventType, pushNotificationGroupQueue.EntityId)

	if err = session.ExecuteBatch(batch); err != nil {
		return nil, errors.WithStack(err)
	}

	return nil, nil
}

func (s *Sender) prepareCustomPushEvents(tokens []database.Device, pushType, kind, title string, body string, headline string,
	key string, customData database.CustomData, userId int64) []notification_gateway.SendPushRequest {
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
				UserId:     userId,
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

// -
func (s *Sender) PushNotification(notification database.Notification, imageUrl string, entityId int64, relatedEntityId int64,
	templateName string, language translation.Language, customKind string, ctx context.Context) (shouldRetry bool, innerErr error) {

	log.Ctx(ctx).Info().
		Str("template_name", templateName).
		Int64("user_id", notification.UserId).
		Int64("entity_id", entityId).
		Int64("related_entity_id", relatedEntityId).
		Msg("[PushNotification] Starting notification processing")

	var template database.RenderTemplate
	var title string
	var body string
	var err error
	isCustomPush := strings.Contains(templateName, "push_admin")

	if !isCustomPush {
		log.Ctx(ctx).Info().Str("template_name", templateName).Msg("[PushNotification] Fetching template from database")

		db := database.GetDbWithContext(database.DbTypeMaster, ctx)
		if err = db.Where("id = ?", templateName).Find(&template).Error; err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("[PushNotification] Failed to fetch template")
			return true, errors.WithStack(err)
		}

		if template.Id != templateName {
			log.Ctx(ctx).Error().Msg("[PushNotification] Template not found")
			return false, errors.WithStack(errors.New("template not found"))
		}

		if template.Muted {
			log.Ctx(ctx).Warn().Msg("[PushNotification] Template is muted, skipping")
			return false, nil
		}
	} else {
		log.Ctx(ctx).Info().Msg("[PushNotification] Using custom push notification")
		title = notification.Title
		body = notification.Message
	}

	template.Id = templateName

	notification.Id = uuid.New()
	notification.CreatedAt = time.Now().UTC()

	if notification.CustomData == nil {
		log.Ctx(ctx).Info().Msg("[PushNotification] Initializing custom data")
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

	log.Ctx(ctx).Debug().Interface("custom_data", notification.CustomData).Msg("[PushNotification] Prepared custom data")

	customDataMarshalled, err := json.Marshal(notification.CustomData)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("[PushNotification] Failed to marshal custom data")
		return false, errors.WithStack(err)
	}

	var notificationInfoMarshalled []byte
	notificationInfoMarshalled, err = json.Marshal(notification)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("[PushNotification] Failed to marshal notification info")
		return false, errors.WithStack(err)
	}

	var renderingVariablesMarshalled []byte
	renderingVariablesMarshalled, err = json.Marshal(notification.RenderingVariables)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("[PushNotification] Failed to marshal rendering variables")
		return false, errors.WithStack(err)
	}

	kind := ""
	if len(customKind) == 0 {
		kind = template.Kind
	} else {
		kind = customKind
	}
	log.Ctx(ctx).Info().Str("kind", kind).Msg("[PushNotification] Notification kind determined")

	session := database.GetScyllaSession()
	batch := session.NewBatch(gocql.UnloggedBatch).WithContext(ctx)

	notificationsCount := int64(1)
	alreadySend := false

	if template.IsGrouped {
		log.Ctx(ctx).Info().Msg("[PushNotification] Processing grouped notifications")

		query := "select user_id, entity_id, related_entity_id from notification_relation where user_id = ? and event_type = ?"
		if relatedEntityId != 0 {
			query = fmt.Sprintf("%v and entity_id = %v", query, entityId)
		}

		notificationRelationIter := session.Query(query, notification.UserId, template.Id).WithContext(ctx).Iter()

		var userIdSelected int64
		var entityIdSelected int64
		var relatedEntityIdSelected int64
		for notificationRelationIter.Scan(&userIdSelected, &entityIdSelected, &relatedEntityIdSelected) {
			if entityIdSelected == entityId && relatedEntityIdSelected == relatedEntityId {
				alreadySend = true
			} else {
				notificationsCount++
			}
		}

		if err = notificationRelationIter.Close(); err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("[PushNotification] Failed to close notification relation iterator")
			return true, errors.WithStack(err)
		}

		batch.Query("update notification_relation set event_applied = true where user_id = ? and event_type = ? "+
			"and entity_id = ? and related_entity_id = ?", notification.UserId, template.Id, entityId, relatedEntityId)

		notificationIter := session.Query("select entity_id, related_entity_id, created_at "+
			"from notification where user_id = ? and event_type = ? and created_at >= ?",
			notification.UserId, template.Id, notification.CreatedAt.Add(-3*24*30*time.Hour)).WithContext(ctx).Iter()

		found := false
		entityIdSelected = 0
		relatedEntityIdSelected = 0
		var createdAt time.Time

		for notificationIter.Scan(&entityIdSelected, &relatedEntityIdSelected, &createdAt) {
			if (relatedEntityId != 0 && entityIdSelected == entityId) || relatedEntityId == 0 {
				found = true
				break
			}
		}

		if err = notificationIter.Close(); err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("[PushNotification] Failed to close notification iterator")
			return true, errors.WithStack(err)
		}

		if found {
			log.Ctx(ctx).Info().
				Int64("entity_id", entityIdSelected).
				Int64("related_entity_id", relatedEntityIdSelected).
				Msg("[PushNotification] Found duplicate notification, deleting")
			batch.Query("delete from notification where user_id = ? and event_type = ? and created_at = ? and entity_id = ? and related_entity_id = ?",
				notification.UserId, template.Id, createdAt, entityIdSelected, relatedEntityIdSelected)
			if err = s.UpdateCreatedAtInGroupQueue(notification.UserId, template.Id, entityIdSelected,
				relatedEntityIdSelected, notification.CreatedAt, ctx); err != nil {
				log.Ctx(ctx).Error().Err(err).Msg("[PushNotification] Failed to update createdAt in group queue")
				return true, errors.WithStack(err)
			}
		}
	}

	var headline string
	var titleMultiple string
	var bodyMultiple string
	var headlineMultiple string

	if notification.RenderingVariables == nil {
		notification.RenderingVariables = database.RenderingVariables{}
	}

	notification.RenderingVariables["notificationsCount"] = strconv.FormatInt(notificationsCount-1, 10)

	if !isCustomPush {
		if template.IsGrouped {
			if firstname, ok := notification.RenderingVariables["firstname"]; ok && len(strings.TrimSpace(firstname)) == 0 {
				if notificationsCount <= 1 {
					return false, nil
				} else {
					notification.RenderingVariables["firstname"] = "Someone"
				}
			}
		}

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

	if notification.Type == "push.profile.following" {
		if notification.Message == "Someone  started following you" {
			notification.Message = "Someone started following you"
		}
	}

	if notification.Type == "push.profile.following" {
		var deletedCount int64
		result := tx.Exec(`
			DELETE FROM notifications 
			WHERE user_id = ? 
			AND related_user_id = ? 
			AND type = ? 
			AND created_at >= ?`,
			notification.UserId,
			notification.RelatedUserId,
			notification.Type,
			time.Now().Add(-24*time.Hour),
		)

		if result.Error != nil {
			log.Ctx(ctx).Error().Err(result.Error).Msg("[PushNotification] Failed to delete existing notifications")
			return true, result.Error
		}

		deletedCount = result.RowsAffected
		if deletedCount > 0 {
			log.Ctx(ctx).Info().
				Int64("deleted_count", deletedCount).
				Int64("user_id", notification.UserId).
				Msg("[PushNotification] Deleted existing notifications")
		}
	}

	if notification.Type == "push.profile.following" || notification.Type == "push.content.like" {
		notification.AggregatedSent = false
	} else {
		notification.AggregatedSent = true
	}
	if err = tx.Create(&notification).Error; err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("[PushNotification] Failed to create notification")
		return true, err
	}

	if err = notificationPkg.IncrementUnreadNotificationsCounter(tx, notification.UserId); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("[PushNotification] Failed to increment unread notifications counter")
		return true, err
	}

	if err = tx.Commit().Error; err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("[PushNotification] Failed to commit transaction")
		return true, errors.WithStack(err)
	}

	if !alreadySend {
		log.Ctx(ctx).Info().
			Int64("user_id", notification.UserId).
			Str("template_id", template.Id).
			Msg("[PushNotification] Sending custom push notification")
		if _, err = s.sendCustomPushTemplateMessageToUser(template.Id, kind, title, body, "", notification.UserId, notification.CustomData, template.IsGrouped,
			entityId, relatedEntityId, notification.CreatedAt, ctx); err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("[PushNotification] Failed to send custom push notification")
			return true, errors.WithStack(err)
		}
	}

	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "notification_id", notification.Id.String())

	isAggregationEligible := notification.Type == "push.profile.following" || notification.Type == "push.content.like"
	isSendDirectly := notification.Type == "push.admin.bulk" || notification.Type == "push.user.after_signup" ||
		notification.Type == "push.user.need.upload" || notification.Type == "push.user.need.avatar"

	if isAggregationEligible || isSendDirectly {
		deviceInfo, err := notificationPkg.GetLatestDeviceForUser(int(notification.UserId), database.GetDbWithContext(database.DbTypeMaster, ctx))
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("[PushNotification] Failed to get token for firebase")
			return false, nil
		}

		if deviceInfo.PushToken != "" {
			if isAggregationEligible {
				log.Ctx(ctx).Info().
					Msg("[PushNotification] Skipping aggregation, recently sent")
				return false, nil
			}
			// Now send the notification (either aggregated or individual)
			data := make(map[string]string)
			for k, v := range notification.CustomData {
				switch value := v.(type) {
				case string:
					data[k] = value
				case int:
					data[k] = fmt.Sprintf("%d", value)
				case int64:
					data[k] = fmt.Sprintf("%d", value)
				case float64:
					data[k] = fmt.Sprintf("%.0f", value)
				case bool:
					data[k] = fmt.Sprintf("%t", value)
				default:
					// Skip other types
				}
			}

			fResp, err := s.firebaseClient.SendNotification(ctx, deviceInfo.PushToken, string(deviceInfo.Platform), "",
				notification.Title, imageUrl, notification.Message, notification.Type, data)
			if err != nil {
				log.Info().Msgf("firebase-reponse fail %v for user-id %v for token %v", fResp, notification.UserId, deviceInfo.PushToken)
				log.Ctx(ctx).Error().Err(err).Msg("[PushNotification] Failed to sent notification on firebase")
				return false, nil
			}
			log.Info().Msgf("firebase-reponse success %v for user-id %v for token %v", fResp, notification.UserId, deviceInfo.PushToken)
			log.Info().Msgf("firebase-reponse %v", fResp)
			log.Info().Msg("Push notification firebase successfully")
		}
	}

	log.Ctx(ctx).Info().
		Str("notification_id", notification.Id.String()).
		Msg("[PushNotification] Notification sent successfully")

	return false, nil
}

func (s *Sender) UpdateCreatedAtInGroupQueue(userId int64, eventType string, entityId int64, relatedEntityId int64,
	newCreatedAt time.Time, ctx context.Context) error {
	session := database.GetScyllaSession()

	deadlineKeys, deadlines := GetDeadlinesForSelect(newCreatedAt)

	query := fmt.Sprintf("select deadline_key, deadline, user_id, event_type, entity_id, created_at, "+
		"notification_count from push_notification_group_queue where deadline_key in (%v) and deadline in (%v) and user_id = ? "+
		"and event_type = ?", utils.JoinDatesForInStatement(deadlineKeys), utils.JoinDatesForInStatement(deadlines))

	if relatedEntityId != 0 {
		query = fmt.Sprintf("%v and entity_id = %v", query, entityId)
	}

	iter := session.Query(query, userId, eventType).WithContext(ctx).Iter()

	pushNotificationGroupQueue := scylla.PushNotificationGroupQueue{}

	batch := session.NewBatch(gocql.UnloggedBatch).WithContext(ctx)

	for iter.Scan(&pushNotificationGroupQueue.DeadlineKey, &pushNotificationGroupQueue.Deadline,
		&pushNotificationGroupQueue.UserId, &pushNotificationGroupQueue.EventType,
		&pushNotificationGroupQueue.EntityId, &pushNotificationGroupQueue.CreatedAt,
		&pushNotificationGroupQueue.NotificationCount) {
		// need to correct select by updated created_at before send push
		batch.Query("update push_notification_group_queue set created_at = ?, notification_count = ? "+
			"where deadline_key = ? and deadline = ? and user_id = ? and event_type = ? and entity_id = ?",
			newCreatedAt, pushNotificationGroupQueue.NotificationCount,
			pushNotificationGroupQueue.DeadlineKey, pushNotificationGroupQueue.Deadline, pushNotificationGroupQueue.UserId,
			pushNotificationGroupQueue.EventType, pushNotificationGroupQueue.EntityId)
	}

	if err := iter.Close(); err != nil {
		return errors.WithStack(err)
	}

	if err := session.ExecuteBatch(batch); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func GetDeadlinesForSelect(fromDate time.Time) (deadlineKeys []time.Time, deadlines []time.Time) {
	deadlineKeysLen := (configs.PushNotificationDeadlineKeyMinutes/configs.PushNotificationDeadlineMinutes)*2 + 1
	deadlineKeys = make([]time.Time, deadlineKeysLen)
	newTime := TimeToNearestMinutes(fromDate, configs.PushNotificationDeadlineKeyMinutes, true)

	for i := 0; i < deadlineKeysLen; i++ {
		deadlineKeys[i] = newTime

		if i != deadlineKeysLen-1 {
			newTime = newTime.Add(configs.PushNotificationDeadlineMinutes * time.Minute)
		}
	}

	deadline := fromDate
	minutesDiff := (deadline.Unix() - TimeToNearestMinutes(deadline, configs.PushNotificationDeadlineMinutes, true).Unix()) / 60
	deadline = deadline.Add(-time.Duration(minutesDiff+configs.PushNotificationDeadlineMinutes) * time.Minute)
	deadline = time.Date(deadline.Year(), deadline.Month(), deadline.Day(), deadline.Hour(), deadline.Minute(), 0, 0, deadline.Location())
	deadlines = []time.Time{deadline, deadline.Add(configs.PushNotificationDeadlineMinutes * time.Minute),
		deadline.Add(2 * configs.PushNotificationDeadlineMinutes * time.Minute)}

	return deadlineKeys, deadlines
}

func (s *Sender) CheckPushNotificationDeadlineMinutes(currentDate time.Time, ctx context.Context) error {
	session := database.GetScyllaSession()

	deadlineKeys, deadlines := GetDeadlinesForSelect(currentDate)

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
	notificationIter.Scan(&userIdFromSelect, &notification.RelatedEntityId, &notification.Title, &notification.Body,
		&notification.Headline, &notification.Kind, &notification.RenderingVariables, &notification.CustomData)

	if err := notificationIter.Close(); err != nil {
		return nil, errors.WithStack(err)
	}

	if userIdFromSelect == 0 { // should do nothing
		return nil, nil
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
	eventType string, createdAt time.Time, entityId int64, notificationCount int64, ctx context.Context) error {
	notification, err := s.getNotificationForGroupSend(userId, eventType, createdAt, entityId, ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	if notification == nil { // should do nothing
		return nil
	}

	var customData database.CustomData
	if err = json.Unmarshal([]byte(notification.CustomData), &customData); err != nil {
		return errors.WithStack(err)
	}

	var renderingVariables database.RenderingVariables
	if err = json.Unmarshal([]byte(notification.RenderingVariables), &renderingVariables); err != nil {
		return errors.WithStack(err)
	}

	if err = s.sendGroupedPush(eventType, notification.Kind, userId, entityId, notificationCount, renderingVariables, customData, ctx); err != nil {
		return errors.WithStack(err)
	}

	if err = s.deleteNotificationFromQueue(deadlineKey, deadline, userId,
		eventType, entityId, ctx); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *Sender) SendDeadlinedNotification(currentDate time.Time, item scylla.PushNotificationGroupQueue, ctx context.Context) (shouldLog bool, innerErr error) {
	flooredCreatedAt := TimeToNearestMinutes(currentDate, configs.PushNotificationDeadlineMinutes, true)

	if !flooredCreatedAt.Before(item.DeadlineKey) && // >=
		flooredCreatedAt.Before(item.DeadlineKey.Add(configs.PushNotificationDeadlineMinutes*time.Minute)) {
		if err := s.updateNotificationQueueAndSendPush(item.DeadlineKey, item.Deadline, item.UserId, item.EventType,
			item.CreatedAt, item.EntityId, item.NotificationCount, ctx); err != nil {
			return true, errors.WithStack(err)
		}

		return true, nil
	}

	ceilDeadline := TimeToNearestMinutes(item.Deadline, configs.PushNotificationDeadlineMinutes, false)
	ceilCurrent := TimeToNearestMinutes(currentDate, configs.PushNotificationDeadlineMinutes, false)

	if !ceilCurrent.After(ceilDeadline) || ceilCurrent.Unix()-ceilDeadline.Unix() > configs.PushNotificationDeadlineMinutes*60 {
		return false, nil
	}

	if err := s.updateNotificationQueueAndSendPush(item.DeadlineKey, item.Deadline, item.UserId, item.EventType,
		item.CreatedAt, item.EntityId, item.NotificationCount, ctx); err != nil {
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

	if err := s.jobber.RegisterTask(string(configs.GeneralPushNotificationTask), func(currentDate string) error {
		apmTransaction := apm_helper.StartNewApmTransaction(string(configs.GeneralPushNotificationTask),
			"push_notification", nil, nil)

		defer func() {
			apmTransaction.End()
		}()

		ctx := boilerplate.CreateCustomContext(context.Background(), apmTransaction, log.Logger)

		var currentDateUnmarshalled time.Time

		if len(currentDate) == 0 {
			currentDateUnmarshalled = time.Now().UTC()
		} else {
			var err error
			currentDateUnmarshalled, err = time.Parse("2006-01-02 15:04:05 -0700 UTC", currentDate)
			if err != nil {
				apm_helper.LogError(errors.WithStack(err), ctx)
				return errors.WithStack(err)
			}
		}

		apm_helper.AddApmLabel(apmTransaction, "current_date", currentDateUnmarshalled)

		if err := s.CheckPushNotificationDeadlineMinutes(currentDateUnmarshalled, ctx); err != nil {
			apm_helper.LogError(errors.WithStack(err), ctx)
			return errors.WithStack(err)
		}

		return nil
	}); err != nil {
		return err
	}

	if err := s.jobber.RegisterPeriodicTask(configs.PushNotificationJobCron,
		string(configs.PeriodicPushNotificationTask), &tasks.Signature{
			Name: string(configs.GeneralPushNotificationTask),
			Args: []tasks.Arg{
				{
					Name:  "currentDate",
					Type:  "string",
					Value: "",
				},
			},
		}); err != nil {
		return err
	}

	return nil
}

func (s *Sender) UnapplyEvent(userId int64, eventType string, entityId int64, relatedEntityId int64, ctx context.Context) error {
	session := database.GetScyllaSession()

	if err := session.Query("update notification_relation set event_applied = false where user_id = ? and "+
		"event_type = ? and entity_id = ? and related_entity_id = ?", userId, eventType, entityId, relatedEntityId).WithContext(ctx).Exec(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *Sender) PushPet2Notification(notification database.Notification, entityId int64, relatedEntityId int64,
	templateName string, language translation.Language, customKind string, ctx context.Context) (shouldRetry bool, innerErr error) {

	log.Ctx(ctx).Info().
		Str("template_name", templateName).
		Int64("user_id", notification.UserId).
		Int64("entity_id", entityId).
		Int64("related_entity_id", relatedEntityId).
		Msg("[PushNotification] Starting notification processing")

	var template database.RenderTemplate
	var title string
	var body string
	var err error
	isCustomPush := strings.Contains(templateName, "push_admin")

	if !isCustomPush {
		log.Ctx(ctx).Info().Str("template_name", templateName).Msg("[PushNotification] Fetching template from database")

		db := database.GetDbWithContext(database.DbTypeMaster, ctx)
		if err = db.Where("id = ?", templateName).Find(&template).Error; err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("[PushNotification] Failed to fetch template")
			return true, errors.WithStack(err)
		}

		if template.Id != templateName {
			log.Ctx(ctx).Error().Msg("[PushNotification] Template not found")
			return false, errors.WithStack(errors.New("template not found"))
		}

		if template.Muted {
			log.Ctx(ctx).Warn().Msg("[PushNotification] Template is muted, skipping")
			return false, nil
		}
	} else {
		log.Ctx(ctx).Info().Msg("[PushNotification] Using custom push notification")
		title = notification.Title
		body = notification.Message
	}

	template.Id = templateName

	notification.Id = uuid.New()
	notification.CreatedAt = time.Now().UTC()

	if notification.CustomData == nil {
		log.Ctx(ctx).Info().Msg("[PushNotification] Initializing custom data")
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

	log.Ctx(ctx).Debug().Interface("custom_data", notification.CustomData).Msg("[PushNotification] Prepared custom data")

	customDataMarshalled, err := json.Marshal(notification.CustomData)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("[PushNotification] Failed to marshal custom data")
		return false, errors.WithStack(err)
	}

	var notificationInfoMarshalled []byte
	notificationInfoMarshalled, err = json.Marshal(notification)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("[PushNotification] Failed to marshal notification info")
		return false, errors.WithStack(err)
	}

	var renderingVariablesMarshalled []byte
	renderingVariablesMarshalled, err = json.Marshal(notification.RenderingVariables)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("[PushNotification] Failed to marshal rendering variables")
		return false, errors.WithStack(err)
	}

	kind := ""
	if len(customKind) == 0 {
		kind = template.Kind
	} else {
		kind = customKind
	}
	log.Ctx(ctx).Info().Str("kind", kind).Msg("[PushNotification] Notification kind determined")

	session := database.GetScyllaSession()
	batch := session.NewBatch(gocql.UnloggedBatch).WithContext(ctx)

	notificationsCount := int64(1)
	alreadySend := false

	if template.IsGrouped {
		log.Ctx(ctx).Info().Msg("[PushNotification] Processing grouped notifications")

		query := "select user_id, entity_id, related_entity_id from notification_relation where user_id = ? and event_type = ?"
		if relatedEntityId != 0 {
			query = fmt.Sprintf("%v and entity_id = %v", query, entityId)
		}

		notificationRelationIter := session.Query(query, notification.UserId, template.Id).WithContext(ctx).Iter()

		var userIdSelected int64
		var entityIdSelected int64
		var relatedEntityIdSelected int64
		for notificationRelationIter.Scan(&userIdSelected, &entityIdSelected, &relatedEntityIdSelected) {
			if entityIdSelected == entityId && relatedEntityIdSelected == relatedEntityId {
				alreadySend = true
			} else {
				notificationsCount++
			}
		}

		if err = notificationRelationIter.Close(); err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("[PushNotification] Failed to close notification relation iterator")
			return true, errors.WithStack(err)
		}

		batch.Query("update notification_relation set event_applied = true where user_id = ? and event_type = ? "+
			"and entity_id = ? and related_entity_id = ?", notification.UserId, template.Id, entityId, relatedEntityId)

		notificationIter := session.Query("select entity_id, related_entity_id, created_at "+
			"from notification where user_id = ? and event_type = ? and created_at >= ?",
			notification.UserId, template.Id, notification.CreatedAt.Add(-3*24*30*time.Hour)).WithContext(ctx).Iter()

		found := false
		entityIdSelected = 0
		relatedEntityIdSelected = 0
		var createdAt time.Time

		for notificationIter.Scan(&entityIdSelected, &relatedEntityIdSelected, &createdAt) {
			if (relatedEntityId != 0 && entityIdSelected == entityId) || relatedEntityId == 0 {
				found = true
				break
			}
		}

		if err = notificationIter.Close(); err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("[PushNotification] Failed to close notification iterator")
			return true, errors.WithStack(err)
		}

		if found {
			log.Ctx(ctx).Info().
				Int64("entity_id", entityIdSelected).
				Int64("related_entity_id", relatedEntityIdSelected).
				Msg("[PushNotification] Found duplicate notification, deleting")
			batch.Query("delete from notification where user_id = ? and event_type = ? and created_at = ? and entity_id = ? and related_entity_id = ?",
				notification.UserId, template.Id, createdAt, entityIdSelected, relatedEntityIdSelected)
			if err = s.UpdateCreatedAtInGroupQueue(notification.UserId, template.Id, entityIdSelected,
				relatedEntityIdSelected, notification.CreatedAt, ctx); err != nil {
				log.Ctx(ctx).Error().Err(err).Msg("[PushNotification] Failed to update createdAt in group queue")
				return true, errors.WithStack(err)
			}
		}
	}

	var headline string
	var titleMultiple string
	var bodyMultiple string
	var headlineMultiple string

	if notification.RenderingVariables == nil {
		notification.RenderingVariables = database.RenderingVariables{}
	}

	notification.RenderingVariables["notificationsCount"] = strconv.FormatInt(notificationsCount-1, 10)

	if !isCustomPush {
		if template.IsGrouped {
			if firstname, ok := notification.RenderingVariables["firstname"]; ok && len(strings.TrimSpace(firstname)) == 0 {
				if notificationsCount <= 1 {
					return false, nil
				} else {
					notification.RenderingVariables["firstname"] = "Someone"
				}
			}
		}

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
		log.Ctx(ctx).Error().Err(err).Msg("[PushNotification] Failed to create notification")
		return true, err
	}

	if err = notificationPkg.IncrementUnreadNotificationsCounter(tx, notification.UserId); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("[PushNotification] Failed to increment unread notifications counter")
		return true, err
	}

	if err = tx.Commit().Error; err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("[PushNotification] Failed to commit transaction")
		return true, errors.WithStack(err)
	}

	if !alreadySend {
		log.Ctx(ctx).Info().
			Int64("user_id", notification.UserId).
			Str("template_id", template.Id).
			Msg("[PushNotification] Sending custom push notification")
		if _, err = s.sendCustomPushTemplateMessageToUser(template.Id, kind, title, body, "", notification.UserId, notification.CustomData, template.IsGrouped,
			entityId, relatedEntityId, notification.CreatedAt, ctx); err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("[PushNotification] Failed to send custom push notification")
			return true, errors.WithStack(err)
		}
	}

	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "notification_id", notification.Id.String())

	log.Ctx(ctx).Info().
		Str("notification_id", notification.Id.String()).
		Msg("[PushNotification] Notification sent successfully")

	return false, nil
}

func (s *Sender) sendPet2GroupedPush(eventType, kind string, userId int64, entityId int64, notificationCount int64,
	renderingVariables database.RenderingVariables, customData database.CustomData, ctx context.Context) error {
	db := database.GetDbWithContext(database.DbTypeReadonly, ctx)

	var template database.RenderTemplate
	if err := db.Where("id = ?", eventType).Find(&template).Error; err != nil {
		return errors.WithStack(err)
	}

	if template.Id != eventType {
		return errors.WithStack(errors.New("template not found"))
	}

	if template.Muted {
		return nil
	}

	userTokens, err := token.GetUserTokens(database.GetDbWithContext(database.DbTypeReadonly, ctx), userId)
	if err != nil {
		return errors.WithStack(err)
	}

	if len(userTokens) == 0 {
		return nil
	}

	isMuted, err := s.settingsService.IsPushNotificationMuted(userId, eventType, ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	if isMuted {
		return nil
	}

	session := database.GetScyllaSession()

	notificationRelationIter := session.Query("select user_id, event_applied from notification_relation where user_id = ? "+
		"and event_type = ? and entity_id = ? and related_entity_id = ?", userId, eventType, entityId, 0).Iter()

	var userIdFromSelect int64
	var eventApplied bool
	notificationRelationIter.Scan(&userIdFromSelect, &eventApplied)

	if err = notificationRelationIter.Close(); err != nil {
		return errors.WithStack(err)
	}

	if userIdFromSelect > 0 && !eventApplied {
		return nil
	}

	var userData user_go.UserRecord

	resp := <-s.userWrapper.GetUsers([]int64{userId}, ctx, false)
	if resp.Error != nil {
		return errors.WithStack(resp.Error.ToError())
	}

	var ok bool
	if userData, ok = resp.Response[userId]; !ok {
		return errors.WithStack(errors.New("user not found"))
	}

	renderingVariables["notificationsCount"] = strconv.FormatInt(notificationCount-1, 10)

	if firstname, ok := renderingVariables["firstname"]; ok && len(strings.TrimSpace(firstname)) == 0 {
		if notificationCount <= 1 {
			return nil
		} else {
			renderingVariables["firstname"] = "Someone"
		}
	}

	var title string
	var body string
	var headline string
	var titleMultiple string
	var bodyMultiple string
	var headlineMultiple string
	title, body, headline, titleMultiple, bodyMultiple, headlineMultiple, err = renderer.Render(template, renderingVariables, userData.Language)
	if err != nil {
		return errors.WithStack(err)
	}

	if notificationCount > 1 {
		title = titleMultiple
		body = bodyMultiple
		headline = headlineMultiple
	}

	if err = <-s.gateway.EnqueuePushForUser(s.prepareCustomPushEvents(userTokens, eventType, kind, title, body, headline, fmt.Sprint(userId), customData,
		userId), ctx); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *Sender) sendCustomPushPet2TemplateMessageToUser(pushType, kind, title, body, headline string, userId int64,
	customData database.CustomData, isGrouped bool, entityId int64, relatedEntityId int64, createdAt time.Time,
	ctx context.Context) (interface{}, error) {
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
		sendResult := <-s.gateway.EnqueuePushForUser(s.prepareCustomPushEvents(userTokens, pushType, kind, title, body, headline, fmt.Sprint(userId), customData,
			userId), ctx)
		return nil, sendResult
	}

	session := database.GetScyllaSession()

	notificationRelationIter := session.Query("select event_applied from notification_relation where user_id = ? "+
		"and event_type = ? and entity_id = ?", userId, pushType, entityId).Iter()

	var eventApplied bool
	notificationRelationIter.Scan(&eventApplied)

	if err = notificationRelationIter.Close(); err != nil {
		return nil, errors.WithStack(err)
	}

	if !eventApplied {
		return nil, nil
	}

	deadlineKeys, deadlines := GetDeadlinesForSelect(createdAt)
	deadline := deadlines[0]

	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "deadline_key", deadlineKeys)
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "deadline", deadlines)

	query := fmt.Sprintf("select deadline_key, deadline, user_id, "+
		"event_type, entity_id, created_at, notification_count from push_notification_group_queue "+
		"where deadline_key in (%v) and deadline in (%v) and user_id = ? and event_type = ?",
		utils.JoinDatesForInStatement(deadlineKeys), utils.JoinDatesForInStatement(deadlines))

	if relatedEntityId != 0 {
		query = fmt.Sprintf("%v and entity_id = %v", query, entityId)
	}

	pushNotificationGroupQueueIter := session.Query(query, userId, pushType).WithContext(ctx).Iter()

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

	flooredCreatedAt := TimeToNearestMinutes(createdAt, configs.PushNotificationDeadlineMinutes, true)

	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "grouped_queued_notifications_count", len(pushNotificationsGroupQueue))
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "push_notifications_group_queue", pushNotificationsGroupQueue)

	for _, item := range pushNotificationsGroupQueue {
		if !flooredCreatedAt.Before(item.DeadlineKey) /* >= */ || item.Deadline.Equal(deadline) {
			continue
		}

		pushNotificationGroupQueue = item
		break
	}

	batch := session.NewBatch(gocql.UnloggedBatch).WithContext(ctx)

	hasPreviousPushNotificationGroupQueueItem := pushNotificationGroupQueue.UserId != 0
	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "has_previous_push_notification_group_queue_item", hasPreviousPushNotificationGroupQueueItem)

	if !hasPreviousPushNotificationGroupQueueItem {
		deadline = TimeToNearestMinutes(createdAt, configs.PushNotificationDeadlineMinutes, true).
			Add(configs.PushNotificationDeadlineMinutes * time.Minute)

		var deadlineKey time.Time
		if configs.PushNotificationDeadlineKeyMinutes != configs.PushNotificationDeadlineMinutes {
			deadlineKey = createdAt.Add(configs.PushNotificationDeadlineKeyMinutes * time.Minute)
			deadlineKey = TimeToNearestMinutes(deadlineKey, configs.PushNotificationDeadlineMinutes, false)
		} else {
			deadlineKey = deadline
		}

		batch.Query("update push_notification_group_queue set created_at = ?, notification_count = ? "+
			"where deadline_key = ? and deadline = ? and user_id = ? and event_type = ? and entity_id = ?",
			createdAt, 1, deadlineKey, deadline, userId, pushType, entityId)

		if err = session.ExecuteBatch(batch); err != nil {
			return nil, errors.WithStack(err)
		}

		return nil, nil
	}

	notificationCount := pushNotificationGroupQueue.NotificationCount + 1

	newDeadline := pushNotificationGroupQueue.Deadline.Add(configs.PushNotificationDeadlineMinutes * time.Minute)

	if notificationCount <= 2 || newDeadline.After(pushNotificationGroupQueue.DeadlineKey) {
		newDeadline = pushNotificationGroupQueue.Deadline
	}

	if relatedEntityId == 0 {
		pushNotificationGroupQueue.EntityId = entityId
		if err = session.Query("delete from push_notification_group_queue where deadline_key = ? and deadline = ? and "+
			"user_id = ? and event_type = ?", pushNotificationGroupQueue.DeadlineKey, pushNotificationGroupQueue.Deadline,
			pushNotificationGroupQueue.UserId, pushNotificationGroupQueue.EventType).Exec(); err != nil {
			return nil, errors.WithStack(err)
		}
	}

	batch.Query("update push_notification_group_queue set created_at = ?, notification_count = ? "+
		"where deadline_key = ? and deadline = ? and user_id = ? and event_type = ? and entity_id = ?",
		pushNotificationGroupQueue.CreatedAt, notificationCount, pushNotificationGroupQueue.DeadlineKey, newDeadline,
		pushNotificationGroupQueue.UserId, pushNotificationGroupQueue.EventType, pushNotificationGroupQueue.EntityId)

	if err = session.ExecuteBatch(batch); err != nil {
		return nil, errors.WithStack(err)
	}

	return nil, nil
}

func (s *Sender) prepareCustomPet2PushEvents(tokens []database.Device, pushType, kind, title string, body string, headline string,
	key string, customData database.CustomData, userId int64) []notification_gateway.SendPushRequest {
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
				UserId:     userId,
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
