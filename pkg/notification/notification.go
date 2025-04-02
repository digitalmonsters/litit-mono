package notification

import (
	"context"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/digitalmonsters/go-common/wrappers/follow"
	"github.com/digitalmonsters/go-common/wrappers/notification_handler"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/database/scylla"
	"github.com/digitalmonsters/notification-handler/pkg/utils"
	"github.com/google/uuid"
	"github.com/pilagod/gorm-cursor-paginator/v2/paginator"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	snappy "github.com/segmentio/kafka-go/compress/snappy/go-xerial-snappy"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
)

func GetNotifications(db *gorm.DB, userId int64, page string, typeGroup TypeGroup, pushAdminSupported bool, limit int,
	userGoWrapper user_go.IUserGoWrapper, followWrapper follow.IFollowWrapper, ctx context.Context) (*NotificationsResponse, error) {
	if strings.Contains(page, "empty") {
		return &NotificationsResponse{
			Data:        make([]NotificationsResponseItem, 0),
			Next:        "empty",
			Prev:        page,
			UnreadCount: 0,
		}, nil
	}

	var pageState []byte

	if len(page) > 0 {
		var err error

		pageState, err = base32.StdEncoding.DecodeString(page)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		pageState, err = snappy.Decode(pageState)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	session := database.GetScyllaSession()

	notificationByTypeGroupView := TypeGroupToScyllaViewName(typeGroup)
	if len(notificationByTypeGroupView) == 0 {
		return nil, errors.WithStack(errors.New("unknown group"))
	}

	if pushAdminSupported {
		notificationByTypeGroupView = fmt.Sprintf("%v_with_push_admin", notificationByTypeGroupView)
	}

	query := fmt.Sprintf("select created_at, event_type, entity_id, related_entity_id from %v where user_id = ?", notificationByTypeGroupView)

	iter := session.Query(query, userId).WithContext(ctx).PageSize(limit).PageState(pageState).Iter()

	nextPageState := iter.PageState()
	scanner := iter.Scanner()

	notifications := make([]database.Notification, 0)
	notificationsCounts := make(map[uuid.UUID]int64)

	for scanner.Next() {
		notificationByTypeGroup := scylla.NotificationByTypeGroup{UserId: userId}

		if err := scanner.Scan(&notificationByTypeGroup.CreatedAt, &notificationByTypeGroup.EventType,
			&notificationByTypeGroup.EntityId, &notificationByTypeGroup.RelatedEntityId); err != nil {
			return nil, errors.WithStack(err)
		}

		notificationIter := session.Query("select title, body, notifications_count, notification_info from notification "+
			"where user_id = ? and event_type = ? and created_at = ? and entity_id = ? and related_entity_id = ?",
			userId, notificationByTypeGroup.EventType, notificationByTypeGroup.CreatedAt,
			notificationByTypeGroup.EntityId, notificationByTypeGroup.RelatedEntityId).Iter()

		var title string
		var body string
		var notificationsCount int64
		var notificationInfo string

		notificationIter.Scan(&title, &body, &notificationsCount, &notificationInfo)

		if err := notificationIter.Close(); err != nil {
			return nil, errors.WithStack(err)
		}

		var notification database.Notification
		if err := json.Unmarshal([]byte(notificationInfo), &notification); err != nil {
			return nil, errors.WithStack(err)
		}

		notification.Title = title
		notification.Message = body
		notificationsCounts[notification.Id] = notificationsCount

		notifications = append(notifications, notification)
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.WithStack(err)
	}
	if err := iter.Close(); err != nil {
		return nil, errors.WithStack(err)
	}

	nextPage := ""

	if len(nextPageState) > 0 {
		nextPage = base32.StdEncoding.EncodeToString(snappy.Encode(nextPageState))
	}

	notificationsResp := mapNotificationsToResponseItems(notifications, notificationsCounts, userGoWrapper, followWrapper, ctx)

	var userNotification database.UserNotification

	if err := db.Where("user_id = ?", userId).Find(&userNotification).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	resp := NotificationsResponse{
		Data:        notificationsResp,
		UnreadCount: userNotification.UnreadCount,
	}

	if len(nextPage) > 0 {
		resp.Next = nextPage
	} else {
		resp.Next = "empty"
	}

	if len(page) > 0 {
		resp.Prev = page
	} else {
		resp.Prev = "empty"
	}

	return &resp, nil
}

func GetNotificationsLegacy(db *gorm.DB, userId int64, page string, typeGroup TypeGroup, pushAdminSupported bool,
	limit int, userGoWrapper user_go.IUserGoWrapper, followWrapper follow.IFollowWrapper, ctx context.Context) (*NotificationsResponse, error) {
	notifications := make([]database.Notification, 0)

	p := paginator.New(
		&paginator.Config{
			Rules: []paginator.Rule{{
				Key:   "CreatedAt",
				Order: paginator.DESC,
			}},
			Limit: limit,
			After: page,
		},
	)

	notificationsTemplates := getNotificationsTemplatesByTypeGroup(typeGroup)

	if !pushAdminSupported {
		notificationsTemplates = lo.Filter(notificationsTemplates, func(item string, i int) bool {
			return item != "push_admin"
		})
	}

	notificationsTypes := make([]string, len(notificationsTemplates))
	for i, templateId := range notificationsTemplates {
		notificationsTypes[i] = database.GetNotificationTypeForAll(templateId)
	}
	notificationsTypes = append(notificationsTypes, "push.intro.successful-upload")

	query := db.Model(notifications).Where("user_id = ?", userId)

	if len(notificationsTypes) > 0 {
		query = query.Where("type in ?", notificationsTypes)
	}

	result, cursor, err := p.Paginate(query, &notifications)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if result.Error != nil {
		return nil, errors.WithStack(result.Error)
	}

	var userNotification database.UserNotification

	if err = db.Where("user_id = ?", userId).Find(&userNotification).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	notificationsResp := mapNotificationsToResponseItems(notifications, nil, userGoWrapper, followWrapper, ctx)

	resp := NotificationsResponse{
		Data:        notificationsResp,
		UnreadCount: userNotification.UnreadCount,
	}

	if cursor.After != nil {
		resp.Next = *cursor.After
	} else {
		resp.Next = "WyIyMDIxLTEyLTIzVDE1OjAwOjEzLjE3N1oiLCI2Zjk4NTljNS1kYmI4LTQyMzMtOWY4Yy1mODM2MTVkODY5MjkiXQ=="
	}

	if cursor.Before != nil {
		resp.Prev = *cursor.Before
	} else {
		resp.Prev = "WyIyMDIxLTEyLTIzVDE1OjAwOjEzLjE3N1oiLCI2Zjk4NTljNS1kYmI4LTQyMzMtOWY4Yy1mODM2MTVkODY5MjkiXQ=="
	}

	return &resp, nil
}

func getNotificationsTemplatesByTypeGroup(typeGroup TypeGroup) []string {
	switch typeGroup {
	case TypeGroupAll:
		return []string{"comment_reply", "comment_vote_like", "comment_vote_dislike", "comment_profile_resource_create",
			"comment_content_resource_create", "follow", "content_posted", "tip", "content_like",
			"bonus_followers", "bonus_time", "content_upload", "spot_upload", "spot_upload_cat", "spot_upload_dog", "content_reject",
			"kyc_status_verified", "kyc_status_rejected", "creator_status_rejected", "creator_status_approved",
			"creator_status_pending", "first_daily_followers_bonus", "first_daily_time_bonus",
			"first_guest_x_earned_points", "first_guest_x_paid_views", "first_x_paid_views",
			"first_referral_joined", "first_video_shared", "first_weekly_followers_bonus", "first_weekly_time_bonus",
			"first_x_paid_views_as_content_owner", "guest_max_earned_points_for_views", "increase_reward_stage_1",
			"increase_reward_stage_2", "registration_verify_bonus", "other_referrals_joined", "custom_reward_increase",
			"megabonus", "first_time_avatar_added", "first_video_uploaded", "first_spot_uploaded", "first_bio_video_uploaded", "add_description_bonus",
			"first_x_paid_views_gender_push", "first_email_marketing_added", "top_daily_spot_bonus", "top_weekly_spot_bonus",
			"last_boring_spots", "first_boring_spots", "warning_boring_spots",
			"monthly_mega_bonus_completed", "monthly_mega_bonus_progress",
			"monthly_mega_bonus_progress_almost_finished", "monthly_mega_bonus_one_day_missing", "monthly_mega_bonus_do_not_miss",
			"push_admin", "first_x_social_media_added", "add_social_subs_target_achieved_bonus", "ads_campaign_rejected", "ads_campaign_approved",
			"music_creator_status_rejected", "music_creator_status_approved", "music_creator_status_pending", "intro",
		}
	case TypeGroupComment:
		return []string{"comment_reply", "comment_vote_like", "comment_vote_dislike", "comment_profile_resource_create",
			"comment_content_resource_create", "intro"}
	case TypeGroupSystem:
		return []string{"push_admin"}
	case TypeGroupFollowing:
		return []string{"follow"}
	default:
		return []string{}
	}
}

func DeleteNotification(db *gorm.DB, userId int64, id uuid.UUID) error {
	var notification database.Notification

	if err := db.Where("id = ?", id).Find(&notification).Error; err != nil {
		return err
	}

	if notification.Id.String() == uuid.Nil.String() || notification.UserId != userId {
		return errors.WithStack(errors.New("notification not found"))
	}

	if err := db.Delete(&notification).Error; err != nil {
		return err
	}

	return nil
}

func ReadAllNotifications(db *gorm.DB, userId int64) error {
	if err := db.Model(&database.UserNotification{}).
		Where("user_id = ?", userId).Update("unread_count", 0).Error; err != nil {
		return err
	}

	return nil
}

func IncrementUnreadNotificationsCounter(db *gorm.DB, userId int64) error {
	if err := db.Exec("insert into user_notifications(user_id, unread_count) values(?, 1) on conflict (user_id) do update set unread_count = user_notifications.unread_count + 1", userId).Error; err != nil {
		return err
	}

	return nil
}

func ListNotificationsByAdmin(db *gorm.DB, req ListNotificationsByAdminRequest, userGoWrapper user_go.IUserGoWrapper,
	followWrapper follow.IFollowWrapper, ctx context.Context) (*ListNotificationsByAdminResponse, error) {
	notifications := make([]database.Notification, 0)
	query := db.Model(notifications)

	var totalCount null.Int
	if req.Offset == 0 {
		var count int64
		if err := query.Count(&count).Error; err != nil {
			return nil, err
		}
		totalCount = null.IntFrom(count)
	}

	if sortingArr := req.Sorting; len(sortingArr) > 0 {
		for _, sorting := range sortingArr {
			sortOrder := " asc"
			if !sorting.IsAscending {
				sortOrder = " desc"
			}
			query = query.Order(sorting.Field + sortOrder)
		}
	}

	if err := query.Offset(req.Offset).Limit(req.Limit).Find(&notifications).Error; err != nil {
		return nil, err
	}

	notificationsResp := mapNotificationsToResponseItems(notifications, nil, userGoWrapper, followWrapper, ctx)

	return &ListNotificationsByAdminResponse{
		Items:      notificationsResp,
		TotalCount: totalCount,
	}, nil
}

func ReadNotification(req ReadNotificationRequest, userId int64, ctx context.Context) error {
	session := database.GetScyllaSession()

	iter := session.Query("select notification_id from user_notifications_read where cluster_key = ? and notification_id = ? and user_id = ? limit 1;",
		GetUserNotificationsReadClusterKey(userId), req.NotificationId, userId).
		WithContext(ctx).Iter()

	isNotificationAlreadyRead := iter.NumRows() > 0

	if err := iter.Close(); err != nil {
		return err
	}

	if !isNotificationAlreadyRead {
		if err := session.Query(
			"insert into user_notifications_read (cluster_key, notification_id, user_id) values (?, ?, ?)",
			GetUserNotificationsReadClusterKey(userId), req.NotificationId, userId,
		).Exec(); err != nil {
			return err
		}

		if err := session.Query(
			"update user_notifications_read_counter set read_count = read_count + ? where notification_id = ?",
			1, req.NotificationId,
		).Exec(); err != nil {
			return err
		}
	}

	return nil
}

func GetNotificationsReadCount(req GetNotificationsReadCountRequest, ctx context.Context) (map[int64]int64, error) {
	notificationsReadCountMap := make(map[int64]int64)

	if len(req.NotificationIds) == 0 {
		return notificationsReadCountMap, nil
	}

	session := database.GetScyllaSession()

	iter := session.Query(fmt.Sprintf("select notification_id, read_count from user_notifications_read_counter where notification_id in (%v);",
		utils.JoinInt64ForInStatement(req.NotificationIds))).WithContext(ctx).Iter()

	var notificationId int64
	var readCount int64

	for iter.Scan(&notificationId, &readCount) {
		notificationsReadCountMap[notificationId] = readCount
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return notificationsReadCountMap, nil
}

func DisableUnregisteredTokens(req notification_handler.DisableUnregisteredTokensRequest, db *gorm.DB) ([]string, error) {
	if err := db.Exec(`delete from "devices" where "pushToken" in ?`, req.Tokens).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return req.Tokens, nil
}

func GetLatestDeviceForUser(userID int, db *gorm.DB) (*database.Device, error) {
	var device database.Device
	if err := db.Raw(`SELECT * FROM devices WHERE "userId" = ? ORDER BY "createdAt" DESC LIMIT 1
	`, userID).Scan(&device).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return &device, nil
}

func CreateNotification(ctx context.Context, req notification_handler.CreateNotificationRequest, db *gorm.DB) (notification_handler.CreateNotificationResponse, error) {

	result := MapInternalToDatabaseNotification(req)

	DeleteUnFollowNotification(ctx, req, db)

	if req.Notifications.Type == "push.profile.following" {
		var deletedCount int64
		result := db.Exec(`
			DELETE FROM notifications 
			WHERE user_id = ? 
			AND related_user_id = ? 
			AND type = ? 
			AND created_at >= ?`,
			req.Notifications.UserID,
			req.Notifications.RelatedUserID,
			req.Notifications.Type,
			time.Now().Add(-24*time.Hour),
		)

		if result.Error != nil {
			log.Ctx(ctx).Error().Err(result.Error).Msg("[PushNotification] Failed to delete existing notifications")
			return notification_handler.CreateNotificationResponse{Status: false}, result.Error
		}

		deletedCount = result.RowsAffected
		if deletedCount > 0 {
			log.Ctx(ctx).Info().
				Int64("deleted_count", deletedCount).
				Int("user_id", req.Notifications.UserID).
				Msg("[PushNotification] Deleted existing notifications")
		}
	}

	if err := db.Create(&result).Error; err != nil {
		return notification_handler.CreateNotificationResponse{Status: false}, err
	}

	if err := IncrementUnreadNotificationsCounter(db, result.UserId); err != nil {
		return notification_handler.CreateNotificationResponse{Status: true}, err
	}

	return notification_handler.CreateNotificationResponse{Status: true}, nil
}

func DeleteUnFollowNotification(ctx context.Context, notification notification_handler.CreateNotificationRequest, db *gorm.DB) error {

	var deletedCount int64
	result := db.Exec(`
		DELETE FROM notifications 
		WHERE user_id = ? 
		AND related_user_id = ? 
		AND type = ? `,
		notification.Notifications.UserID,
		notification.Notifications.RelatedUserID,
		"push.profile.following",
	)

	if result.Error != nil {
		log.Ctx(ctx).Error().Err(result.Error).Msg("[PushNotification] Failed to delete existing notifications")
		return result.Error
	}

	deletedCount = result.RowsAffected
	if deletedCount > 0 {
		log.Ctx(ctx).Info().
			Int64("deleted_count", deletedCount).
			Int64("user_id", int64(notification.Notifications.UserID)).
			Msg("[PushNotification] Deleted existing notifications")
	}

	return nil
}

func DeleteNotificationByIntroIDAndType(ctx context.Context, db *gorm.DB, introID int) error {
	result := db.Table("notifications").
		Where("custom_data->>'intro_id' = ?", fmt.Sprintf("%d", introID)).
		Where("type = ?", "push.intro.successful-upload").
		Delete(nil)

	if result.Error != nil {
		return fmt.Errorf("failed to delete notification: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		log.Error().Msgf("no notification found with intro_id: %d ", introID)
	}

	return nil
}

func MapInternalToDatabaseNotification(req notification_handler.CreateNotificationRequest) database.Notification {
	return database.Notification{
		Id:                 uuid.New(), // Generating a new UUID
		UserId:             int64(req.Notifications.UserID),
		Type:               req.Notifications.Type,
		Title:              req.Notifications.Title,
		Message:            req.Notifications.Message,
		RelatedUserId:      ToNullInt(req.Notifications.RelatedUserID),
		CommentId:          ToNullInt(req.Notifications.CommentID),
		ContentId:          ToNullInt(req.Notifications.ContentID),
		Content:            ToNotificationContent(req.Notifications.Content),
		QuestionId:         ToNullInt(req.Notifications.QuestionID),
		CreatedAt:          req.Notifications.CreatedAt,
		RenderingVariables: ToRenderingVariables(req.Notifications.RenderingVariables),
		CustomData:         req.Notifications.CustomData,
	}
}

func GetInAppNotifications(userID int64, db *gorm.DB) ([]database.InAppNotification, error) {
	var notifications []database.InAppNotification

	tx := db.Begin()

	if err := tx.Where("user_id = ? AND is_shown = ?", userID, false).
		Order("created_at DESC").
		Find(&notifications).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if len(notifications) == 0 {
		tx.Rollback()
		return notifications, nil
	}

	var ids []uuid.UUID
	for _, notification := range notifications {
		ids = append(ids, notification.Id)
	}

	if err := tx.Model(&database.InAppNotification{}).
		Where("id IN ?", ids).
		Update("is_shown", true).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return notifications, nil
}

func CreateInAppNotification(req notification_handler.CreateNotificationRequest, db *gorm.DB) error {
	inAppNotification := MapInternalToDatabaseInAppNotification(req)
	if err := db.Create(&inAppNotification).Error; err != nil {
		return err
	}
	return nil
}

func MapInternalToDatabaseInAppNotification(req notification_handler.CreateNotificationRequest) database.InAppNotification {
	return database.InAppNotification{
		Id:                 uuid.New(), // Generating a new UUID
		UserId:             int64(req.Notifications.UserID),
		Type:               req.Notifications.Type,
		Title:              req.Notifications.Title,
		Message:            req.Notifications.Message,
		RelatedUserId:      ToNullInt(req.Notifications.RelatedUserID),
		CommentId:          ToNullInt(req.Notifications.CommentID),
		ContentId:          ToNullInt(req.Notifications.ContentID),
		Content:            ToNotificationContent(req.Notifications.Content),
		QuestionId:         ToNullInt(req.Notifications.QuestionID),
		CreatedAt:          req.Notifications.CreatedAt,
		RenderingVariables: ToRenderingVariables(req.Notifications.RenderingVariables),
		CustomData:         req.Notifications.CustomData,
		IsShown:            false,
	}
}

func ToNullInt(value *int) null.Int {
	if value != nil {
		return null.IntFrom(int64(*value))
	}
	return null.Int{}
}

func ToNotificationContent(data map[string]interface{}) *database.NotificationContent {
	if data == nil {
		return nil
	}
	return &database.NotificationContent{
		Id:      int64(data["id"].(float64)), // Adjust casting based on data type
		Width:   int(data["width"].(float64)),
		Height:  int(data["height"].(float64)),
		VideoId: data["video_id"].(string),
	}
}

func ToRenderingVariables(data map[string]interface{}) database.RenderingVariables {
	renderingVars := make(database.RenderingVariables)
	for key, value := range data {
		if strValue, ok := value.(string); ok {
			renderingVars[key] = strValue
		}
	}
	return renderingVars
}

func GetPet2Notifications(db *gorm.DB, userId int64, page string, typeGroup TypeGroup, pushAdminSupported bool, limit int,
	userGoWrapper user_go.IUserGoWrapper, followWrapper follow.IFollowWrapper, ctx context.Context) (*NotificationsResponse, error) {
	if strings.Contains(page, "empty") {
		return &NotificationsResponse{
			Data:        make([]NotificationsResponseItem, 0),
			Next:        "empty",
			Prev:        page,
			UnreadCount: 0,
		}, nil
	}

	var pageState []byte

	if len(page) > 0 {
		var err error

		pageState, err = base32.StdEncoding.DecodeString(page)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		pageState, err = snappy.Decode(pageState)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	session := database.GetScyllaSession()

	notificationByTypeGroupView := TypeGroupToScyllaViewName(typeGroup)
	if len(notificationByTypeGroupView) == 0 {
		return nil, errors.WithStack(errors.New("unknown group"))
	}

	if pushAdminSupported {
		notificationByTypeGroupView = fmt.Sprintf("%v_with_push_admin", notificationByTypeGroupView)
	}

	query := fmt.Sprintf("select created_at, event_type, entity_id, related_entity_id from %v where user_id = ?", notificationByTypeGroupView)

	iter := session.Query(query, userId).WithContext(ctx).PageSize(limit).PageState(pageState).Iter()

	nextPageState := iter.PageState()
	scanner := iter.Scanner()

	notifications := make([]database.Notification, 0)
	notificationsCounts := make(map[uuid.UUID]int64)

	for scanner.Next() {
		notificationByTypeGroup := scylla.NotificationByTypeGroup{UserId: userId}

		if err := scanner.Scan(&notificationByTypeGroup.CreatedAt, &notificationByTypeGroup.EventType,
			&notificationByTypeGroup.EntityId, &notificationByTypeGroup.RelatedEntityId); err != nil {
			return nil, errors.WithStack(err)
		}

		notificationIter := session.Query("select title, body, notifications_count, notification_info from notification "+
			"where user_id = ? and event_type = ? and created_at = ? and entity_id = ? and related_entity_id = ?",
			userId, notificationByTypeGroup.EventType, notificationByTypeGroup.CreatedAt,
			notificationByTypeGroup.EntityId, notificationByTypeGroup.RelatedEntityId).Iter()

		var title string
		var body string
		var notificationsCount int64
		var notificationInfo string

		notificationIter.Scan(&title, &body, &notificationsCount, &notificationInfo)

		if err := notificationIter.Close(); err != nil {
			return nil, errors.WithStack(err)
		}

		var notification database.Notification
		if err := json.Unmarshal([]byte(notificationInfo), &notification); err != nil {
			return nil, errors.WithStack(err)
		}

		notification.Title = title
		notification.Message = body
		notificationsCounts[notification.Id] = notificationsCount

		notifications = append(notifications, notification)
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.WithStack(err)
	}
	if err := iter.Close(); err != nil {
		return nil, errors.WithStack(err)
	}

	nextPage := ""

	if len(nextPageState) > 0 {
		nextPage = base32.StdEncoding.EncodeToString(snappy.Encode(nextPageState))
	}

	notificationsResp := mapNotificationsToResponseItems(notifications, notificationsCounts, userGoWrapper, followWrapper, ctx)

	var userNotification database.UserNotification

	if err := db.Where("user_id = ?", userId).Find(&userNotification).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	resp := NotificationsResponse{
		Data:        notificationsResp,
		UnreadCount: userNotification.UnreadCount,
	}

	if len(nextPage) > 0 {
		resp.Next = nextPage
	} else {
		resp.Next = "empty"
	}

	if len(page) > 0 {
		resp.Prev = page
	} else {
		resp.Prev = "empty"
	}

	return &resp, nil
}

func GetPet2NotificationsLegacy(db *gorm.DB, userId int64, page string, typeGroup TypeGroup, pushAdminSupported bool,
	limit int, userGoWrapper user_go.IUserGoWrapper, followWrapper follow.IFollowWrapper, ctx context.Context) (*NotificationsResponse, error) {
	notifications := make([]database.Notification, 0)

	p := paginator.New(
		&paginator.Config{
			Rules: []paginator.Rule{{
				Key:   "CreatedAt",
				Order: paginator.DESC,
			}},
			Limit: limit,
			After: page,
		},
	)

	notificationsTemplates := getNotificationsTemplatesByTypeGroup(typeGroup)

	if !pushAdminSupported {
		notificationsTemplates = lo.Filter(notificationsTemplates, func(item string, i int) bool {
			return item != "push_admin"
		})
	}

	notificationsTypes := make([]string, len(notificationsTemplates))
	for i, templateId := range notificationsTemplates {
		notificationsTypes[i] = database.GetNotificationTypeForAll(templateId)
	}

	query := db.Model(notifications).Where("user_id = ?", userId)

	if len(notificationsTypes) > 0 {
		query = query.Where("type in ?", notificationsTypes)
	}

	result, cursor, err := p.Paginate(query, &notifications)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if result.Error != nil {
		return nil, errors.WithStack(result.Error)
	}

	var userNotification database.UserNotification

	if err = db.Where("user_id = ?", userId).Find(&userNotification).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	notificationsResp := mapNotificationsToResponseItems(notifications, nil, userGoWrapper, followWrapper, ctx)

	resp := NotificationsResponse{
		Data:        notificationsResp,
		UnreadCount: userNotification.UnreadCount,
	}

	if cursor.After != nil {
		resp.Next = *cursor.After
	} else {
		resp.Next = "WyIyMDIxLTEyLTIzVDE1OjAwOjEzLjE3N1oiLCI2Zjk4NTljNS1kYmI4LTQyMzMtOWY4Yy1mODM2MTVkODY5MjkiXQ=="
	}

	if cursor.Before != nil {
		resp.Prev = *cursor.Before
	} else {
		resp.Prev = "WyIyMDIxLTEyLTIzVDE1OjAwOjEzLjE3N1oiLCI2Zjk4NTljNS1kYmI4LTQyMzMtOWY4Yy1mODM2MTVkODY5MjkiXQ=="
	}

	return &resp, nil
}

func GetDeviceTokensByUserID(userID []int64, db *gorm.DB) (GetDeviceTokensRPCResponse, error) {

	var devices []database.Device
	query := `
		SELECT DISTINCT ON ("userId") *
		FROM devices
		WHERE "userId" IN ?
		ORDER BY "userId", "createdAt" DESC
	`
	if err := db.Raw(query, userID).Scan(&devices).Error; err != nil {
		log.Err(err).Msg("Failed to fetch latest devices per user")
		return GetDeviceTokensRPCResponse{}, err
	}

	if len(devices) == 0 {
		return GetDeviceTokensRPCResponse{}, errors.WithStack(errors.New("no devices found"))
	}

	tokens := make(map[int64]string)

	for _, device := range devices {
		tokens[device.UserId] = device.PushToken
	}

	response := GetDeviceTokensRPCResponse{
		DeviceTokens: tokens,
	}

	return response, nil
}
