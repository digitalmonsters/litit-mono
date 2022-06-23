package migrator

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/RichardKnop/machinery/v1"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/translation"
	"github.com/digitalmonsters/notification-handler/configs"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/database/scylla"
	"github.com/digitalmonsters/notification-handler/pkg/renderer"
	"github.com/gocql/gocql"
	"github.com/pilagod/gorm-cursor-paginator/v2/paginator"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.elastic.co/apm"
	"go.elastic.co/apm/module/apmhttp"
	"strconv"
	"strings"
	"time"
)

func RegisterMigratorTasks(jobber *machinery.Server) error {
	if err := jobber.RegisterTask(string(configs.Migrator1Task),
		func(traceHeader string) error {
			var apmTransaction *apm.Transaction

			if parsed, err := apmhttp.ParseTraceparentHeader(traceHeader); err != nil {
				log.Err(err).Send()
				apmTransaction = apm_helper.StartNewApmTransaction(string(configs.Migrator1Task),
					"migrator", nil, nil)
			} else {
				apmTransaction = apm_helper.StartNewApmTransactionWithTraceData(string(configs.Migrator1Task),
					"migrator", nil, parsed)
			}

			defer func() {
				apmTransaction.End()
			}()

			ctx := boilerplate.CreateCustomContext(context.Background(), apmTransaction, log.Logger)

			if err := MigrateNotificationsToScylla(ctx); err != nil {
				apm_helper.LogError(errors.WithStack(err), ctx)
				return errors.WithStack(err)
			}

			return nil
		}); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func MigrateNotificationsToScylla(ctx context.Context) error {
	logger := zerolog.Ctx(ctx)

	logger.Info().Msg("[MigrateNotificationsToScylla] start")

	db := database.GetDbWithContext(database.DbTypeReadonly, ctx)
	timeNow := time.Now().UTC()

	var templates []database.RenderTemplate
	if err := db.Find(&templates).Error; err != nil {
		return errors.WithStack(err)
	}

	if len(templates) == 0 {
		logger.Info().Msg("[MigrateNotificationsToScylla] templates len 0")
		return nil
	}

	templatesMap := make(map[string]database.RenderTemplate, len(templates))
	for _, template := range templates {
		templatesMap[template.Id] = template
	}

	session := database.GetScyllaSession()
	notificationRelationBatch := session.NewBatch(gocql.UnloggedBatch).WithContext(ctx)
	batchCount := 0
	maxBatchCount := 100

	groupedNotifications := make(map[int64]map[string]map[int64]scylla.Notification)
	var scyllaNotificationsToUpdate []scylla.Notification

	after := ""
	isFirst := true
	for {
		if !isFirst && len(after) == 0 {
			break
		}

		isFirst = false

		p := paginator.New(
			&paginator.Config{
				Rules: []paginator.Rule{
					{
						Key:   "UserId",
						Order: paginator.ASC,
					},
					{
						Key:   "CreatedAt",
						Order: paginator.ASC,
					},
				},
				Limit: 1000,
			},
		)

		if len(after) > 0 {
			p.SetAfterCursor(after)
		}

		var dbNotifications []database.Notification
		query := db.Where("created_at >= ? and created_at <= ?", timeNow.Add(-3*24*30*time.Hour),
			time.Date(2022, 6, 18, 10, 25, 0, 0, time.UTC))
		result, cursor, err := p.Paginate(query, &dbNotifications)
		if err != nil {
			return errors.WithStack(err)
		}

		if result.Error != nil {
			return errors.WithStack(result.Error)
		}

		if len(dbNotifications) == 0 {
			break
		}

		if cursor.After != nil {
			after = *cursor.After
		} else {
			after = ""
		}

		logger.Info().Msgf("[MigrateNotificationsToScylla] before dbNotifications iterations, len %v", len(dbNotifications))

		for _, dbNotification := range dbNotifications {
			if dbNotification.Type == "push.admin.bulk" || dbNotification.Type == "popup" || len(dbNotification.Type) == 0 || dbNotification.Type == "push.bonus.time" {
				continue
			}

			eventTypes := database.GetNotificationTemplates(dbNotification.Type)

			if len(eventTypes) == 0 {
				logger.Info().Msgf("[MigrateNotificationsToScylla] eventTypes for \"%v\" not found", dbNotification.Type)
				continue
			}

			eventType := ""

			if len(eventTypes) > 1 {
				switch dbNotification.Type {
				case "push.paid_views.first":
					if strings.Contains(dbNotification.Title, "Create your account") {
						eventType = "first_guest_x_paid_views"
					} else {
						eventType = "first_x_paid_views"
					}
				case "push.guest.after_install":
					if strings.Contains(dbNotification.Message, "Complete your account creation") {
						eventType = "guest_after_install_first_push"
					} else if strings.Contains(dbNotification.Message, "On Lit.it the more viral videos you watch the more your earn") {
						eventType = "guest_after_install_second_push"
					} else {
						eventType = "guest_after_install_third_push"
					}
				case "push.user.after_signup":
					if strings.Contains(dbNotification.Message, "Few things you need to know about Lit.it") {
						eventType = "user_after_signup_first_push"
					} else if strings.Contains(dbNotification.Message, "Check who earned the most for inviting friends to Lit.it") {
						eventType = "user_after_signup_second_push"
					} else if strings.Contains(dbNotification.Message, "Who earned the most LIT points? Check this out") {
						eventType = "user_after_signup_third_push"
					} else if strings.Contains(dbNotification.Message, "Check out TOP viral videos on Lit.it & earn LIT points") {
						eventType = "user_after_signup_fourth_push"
					} else {
						eventType = "user_after_signup_fifth_push"
					}
				case "push.comment.vote":
					if strings.Contains(dbNotification.Message, "disliked your comment") {
						eventType = "comment_vote_dislike"
					} else {
						eventType = "comment_vote_like"
					}
				case "push.kyc.status":
					if strings.Contains(dbNotification.Message, "Your identity verification has been approved") {
						eventType = "kyc_status_verified"
					} else {
						eventType = "kyc_status_rejected"
					}
				case "push.content-creator.status":
					if strings.Contains(dbNotification.Message, "Your Creator approval process has been rejected.") {
						eventType = "creator_status_rejected"
					} else if strings.Contains(dbNotification.Message, "Your Creator status has been approved") {
						eventType = "creator_status_approved"
					} else {
						eventType = "creator_status_pending"
					}
				}
			} else {
				eventType = eventTypes[0]
			}

			if len(eventType) == 0 {
				logger.Info().Msgf("[MigrateNotificationsToScylla] eventTypes for \"%v\" not found is switch", dbNotification.Type)
				continue
			}

			template, ok := templatesMap[eventType]
			if !ok {
				logger.Info().Msgf("[MigrateNotificationsToScylla] template for \"%v\" not found", eventType)
				continue
			}

			customDataMarshalled, _ := json.Marshal(dbNotification.CustomData)
			if len(customDataMarshalled) == 0 {
				customDataMarshalled = []byte("{}")
			}

			renderingVariablesMarshalled, _ := json.Marshal(dbNotification.RenderingVariables)
			if len(renderingVariablesMarshalled) == 0 {
				renderingVariablesMarshalled = []byte("{}")
			}

			notificationMarshalled, _ := json.Marshal(dbNotification)
			if len(notificationMarshalled) == 0 {
				notificationMarshalled = []byte("{}")
			}

			scyllaNotification := scylla.Notification{
				UserId:             dbNotification.UserId,
				EventType:          eventType,
				CreatedAt:          dbNotification.CreatedAt,
				EntityId:           dbNotification.UserId,
				CustomData:         string(customDataMarshalled),
				RenderingVariables: string(renderingVariablesMarshalled),
				NotificationInfo:   string(notificationMarshalled),
				Kind:               template.Kind,
				NotificationsCount: 1,
				Title:              dbNotification.Title,
				Body:               dbNotification.Message,
			}

			switch eventType {
			case "comment_reply":
				scyllaNotification.Kind = "default"
				scyllaNotification.RelatedEntityId = dbNotification.RelatedUserId.ValueOrZero()

				if dbNotification.Comment != nil {
					scyllaNotification.EntityId = dbNotification.Comment.ParentId.ValueOrZero()
				}
			case "comment_content_resource_create":
				scyllaNotification.Kind = "default"
				scyllaNotification.EntityId = dbNotification.ContentId.ValueOrZero()
				scyllaNotification.RelatedEntityId = dbNotification.RelatedUserId.ValueOrZero()
			case "comment_profile_resource_create":
				scyllaNotification.Kind = "default"
				scyllaNotification.EntityId = dbNotification.RelatedUserId.ValueOrZero()
			case "spot_upload":
				fallthrough
			case "content_upload":
				scyllaNotification.Kind = "default"
				scyllaNotification.EntityId = dbNotification.ContentId.ValueOrZero()
				scyllaNotification.RelatedEntityId = dbNotification.UserId
			case "creator_status_rejected":
				fallthrough
			case "creator_status_approved":
				fallthrough
			case "creator_status_pending":
				scyllaNotification.Kind = "content_creator"
				scyllaNotification.EntityId = dbNotification.UserId
			case "follow":
				scyllaNotification.Kind = "user_follow"
				scyllaNotification.EntityId = dbNotification.RelatedUserId.ValueOrZero()
			case "kyc_status_verified":
				fallthrough
			case "kyc_status_rejected":
				scyllaNotification.Kind = "default"
				scyllaNotification.EntityId = dbNotification.UserId
			case "content_like":
				scyllaNotification.Kind = "default"
				scyllaNotification.EntityId = dbNotification.ContentId.ValueOrZero()
				scyllaNotification.RelatedEntityId = dbNotification.RelatedUserId.ValueOrZero()
			case "push_admin":
				scyllaNotification.Kind = "default"
				scyllaNotification.EntityId = dbNotification.UserId
			case "tip":
				fallthrough
			case "bonus_time":
				fallthrough
			case "bonus_followers":
				scyllaNotification.Kind = "default"
				scyllaNotification.EntityId = dbNotification.RelatedUserId.ValueOrZero()
			case "comment_vote_dislike":
				fallthrough
			case "comment_vote_like":
				scyllaNotification.Kind = "default"
				scyllaNotification.EntityId = dbNotification.CommentId.ValueOrZero()
				scyllaNotification.RelatedEntityId = dbNotification.RelatedUserId.ValueOrZero()
			}

			if template.IsGrouped {
				groupedNotification, ok := groupedNotifications[dbNotification.UserId]
				if !ok {
					groupedNotifications[dbNotification.UserId] = map[string]map[int64]scylla.Notification{
						eventType: {
							scyllaNotification.EntityId: scyllaNotification,
						},
					}
				} else {
					groupedTemplateNotification, ok := groupedNotification[eventType]
					if !ok {
						groupedNotifications[dbNotification.UserId][eventType] = map[int64]scylla.Notification{
							scyllaNotification.EntityId: scyllaNotification,
						}
					} else {
						groupedTemplateNotificationEntity, ok := groupedTemplateNotification[scyllaNotification.EntityId]
						if !ok {
							groupedNotifications[dbNotification.UserId][eventType][scyllaNotification.EntityId] = scyllaNotification
						} else {
							scyllaNotification.NotificationsCount = groupedTemplateNotificationEntity.NotificationsCount + 1
							groupedNotifications[dbNotification.UserId][eventType][scyllaNotification.EntityId] = scyllaNotification
						}
					}
				}
			}

			ttl := timeNow.Unix() - scyllaNotification.CreatedAt.Unix()

			if ttl <= 0 {
				ttl = 7776000
			}

			if !template.IsGrouped {
				scyllaNotificationsToUpdate = append(scyllaNotificationsToUpdate, scyllaNotification)
			} else {
				notificationRelationBatch.Query(fmt.Sprintf("update notification_relation using ttl %v set event_applied = true "+
					"where user_id = ? and event_type = ? and entity_id = ? and related_entity_id = ?", ttl),
					scyllaNotification.UserId, scyllaNotification.EventType, scyllaNotification.EntityId,
					scyllaNotification.RelatedEntityId)
				batchCount++

				if batchCount == maxBatchCount {
					logger.Info().Msgf("[MigrateNotificationsToScylla] before notificationRelationBatch Execute, len %v", batchCount)
					if err = session.ExecuteBatch(notificationRelationBatch); err != nil {
						logger.Error().Msgf("[MigrateNotificationsToScylla] ExecuteBatch err %v", err.Error())
						return errors.WithStack(err)
					}
					batchCount = 0
					notificationRelationBatch = session.NewBatch(gocql.UnloggedBatch).WithContext(ctx)
					logger.Info().Msg("[MigrateNotificationsToScylla] after notificationRelationBatch Execute")
				}
			}
		}
	}

	logger.Info().Msg("[MigrateNotificationsToScylla] after dbNotifications iterations")

	if batchCount != 0 {
		logger.Info().Msgf("[MigrateNotificationsToScylla] before notificationRelationBatch Execute, len %v", batchCount)
		if err := session.ExecuteBatch(notificationRelationBatch); err != nil {
			logger.Error().Msgf("[MigrateNotificationsToScylla] notificationRelationBatch ExecuteBatch err %v", err.Error())
			return errors.WithStack(err)
		}
		logger.Info().Msg("[MigrateNotificationsToScylla] after notificationRelationBatch Execute")
	}

	logger.Info().Msgf("[MigrateNotificationsToScylla] before groupedNotifications iterations, len %v", len(groupedNotifications))

	for _, scyllaNotificationsEntities := range groupedNotifications {
		for _, scyllaNotificationEntity := range scyllaNotificationsEntities {
			for _, scyllaNotification := range scyllaNotificationEntity {
				query := "select entity_id, related_entity_id, created_at, notifications_count, " +
					"title, body, headline, kind, rendering_variables, custom_data, notification_info from notification where " +
					"user_id = ? and event_type = ? and created_at > ?"

				if scyllaNotification.RelatedEntityId == 0 {
					query = fmt.Sprintf("%v limit 1", query)
				}

				iter := session.Query(query, scyllaNotification.UserId, scyllaNotification.EventType, scyllaNotification.CreatedAt).Iter()

				var notifications []scylla.Notification
				var notification scylla.Notification
				for iter.Scan(&notification.EntityId, &notification.RelatedEntityId, &notification.CreatedAt,
					&notification.NotificationsCount, &notification.Title, &notification.Body, &notification.Headline,
					&notification.Kind, &notification.RenderingVariables, &notification.CustomData,
					&notification.NotificationInfo) {
					notification.UserId = scyllaNotification.UserId
					notification.EventType = scyllaNotification.EventType
					notifications = append(notifications, notification)
				}

				if err := iter.Close(); err != nil {
					return errors.WithStack(err)
				}

				template, ok := templatesMap[scyllaNotification.EventType]
				if !ok {
					logger.Info().Msgf("[MigrateNotificationsToScylla] template \"%v\" not found in templatesMap", scyllaNotification.EventType)
					continue
				}

				if len(notifications) != 0 {
					notification = scylla.Notification{}

					if scyllaNotification.RelatedEntityId == 0 {
						notification = notifications[0]
					} else {
						for _, v := range notifications {
							if v.EntityId == scyllaNotification.EntityId {
								notification = v
								break
							}
						}
					}

					notificationsCount := scyllaNotification.NotificationsCount + notification.NotificationsCount
					scyllaNotification = notification
					scyllaNotification.NotificationsCount = notificationsCount
				}

				if len(scyllaNotification.RenderingVariables) == 0 {
					scyllaNotification.RenderingVariables = "{}"
				}

				var renderingVariables database.RenderingVariables
				if err := json.Unmarshal([]byte(scyllaNotification.RenderingVariables), &renderingVariables); err != nil {
					continue
				}

				if renderingVariables == nil {
					renderingVariables = database.RenderingVariables{}
				}

				renderingVariables["notificationsCount"] = strconv.FormatInt(scyllaNotification.NotificationsCount, 10)

				renderingVariablesMarshalled, _ := json.Marshal(renderingVariables)
				if len(renderingVariablesMarshalled) == 0 {
					renderingVariablesMarshalled = []byte("{}")
				}

				scyllaNotification.RenderingVariables = string(renderingVariablesMarshalled)

				title, body, headline, titleMultiple, bodyMultiple, headlineMultiple, err :=
					renderer.Render(template, renderingVariables, translation.DefaultUserLanguage)
				if err != nil {
					continue
				}

				if scyllaNotification.NotificationsCount == 1 {
					scyllaNotification.Title = title
					scyllaNotification.Body = body
					scyllaNotification.Headline = headline
				} else {
					scyllaNotification.Title = titleMultiple
					scyllaNotification.Body = bodyMultiple
					scyllaNotification.Headline = headlineMultiple
				}

				scyllaNotificationsToUpdate = append(scyllaNotificationsToUpdate, scyllaNotification)
			}
		}
	}

	logger.Info().Msgf("[MigrateNotificationsToScylla] before scyllaNotificationsToUpdate iterations, len %v", len(scyllaNotificationsToUpdate))

	notificationBatch := session.NewBatch(gocql.UnloggedBatch).WithContext(ctx)
	batchCount = 0

	for _, scyllaNotification := range scyllaNotificationsToUpdate {
		ttl := timeNow.Unix() - scyllaNotification.CreatedAt.Unix()
		if ttl <= 0 {
			ttl = 7776000
		}

		notificationBatch.Query(fmt.Sprintf("update notification using ttl %v set notifications_count = ?, title = ?, "+
			"body = ?, headline = ?, kind = ?, rendering_variables = ?, custom_data = ?, notification_info = ? "+
			"where user_id = ? and event_type = ? and created_at = ? and entity_id = ? and related_entity_id = ?",
			ttl), scyllaNotification.NotificationsCount, scyllaNotification.Title, scyllaNotification.Body,
			scyllaNotification.Headline, scyllaNotification.Kind, scyllaNotification.RenderingVariables,
			scyllaNotification.CustomData, scyllaNotification.NotificationInfo, scyllaNotification.UserId,
			scyllaNotification.EventType, scyllaNotification.CreatedAt, scyllaNotification.EntityId,
			scyllaNotification.RelatedEntityId,
		)
		batchCount++

		if batchCount == maxBatchCount {
			logger.Info().Msgf("[MigrateNotificationsToScylla] before notificationBatch Execute, len %v", batchCount)
			if err := session.ExecuteBatch(notificationBatch); err != nil {
				logger.Error().Msgf("[MigrateNotificationsToScylla] ExecuteBatch err %v", err.Error())
				return errors.WithStack(err)
			}
			batchCount = 0
			notificationBatch = session.NewBatch(gocql.UnloggedBatch).WithContext(ctx)
			logger.Info().Msg("[MigrateNotificationsToScylla] after notificationBatch Execute")
		}
	}

	if batchCount != 0 {
		logger.Info().Msgf("[MigrateNotificationsToScylla] before notificationBatch Execute, len %v", batchCount)
		if err := session.ExecuteBatch(notificationBatch); err != nil {
			logger.Error().Msgf("[MigrateNotificationsToScylla] notificationBatch ExecuteBatch err %v", err.Error())
			return errors.WithStack(err)
		}
		logger.Info().Msg("[MigrateNotificationsToScylla] after notificationBatch Execute")
	}

	logger.Info().Msg("[MigrateNotificationsToScylla] end")

	return nil
}
