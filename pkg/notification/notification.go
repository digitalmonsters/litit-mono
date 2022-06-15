package notification

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/wrappers/follow"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"strings"
)

func GetNotifications(db *gorm.DB, userId int64, page string, typeGroup TypeGroup, userGoWrapper user_go.IUserGoWrapper,
	followWrapper follow.IFollowWrapper, ctx context.Context) (*NotificationsResponse, error) {
	var pageState []byte

	if len(page) > 0 {
		if strings.Contains(page, "empty") {
			page = ""
		} else {
			var err error
			pageState, err = base64.StdEncoding.DecodeString(page)
			if err != nil {
				return nil, errors.WithStack(err)
			}
		}
	}

	templates := make([]string, 0)
	for _, notificationType := range getNotificationsTypesByTypeGroup(typeGroup) {
		templates = append(templates, database.GetNotificationTemplates(notificationType)...)
	}

	session := database.GetScyllaSession()
	iter := session.Query(fmt.Sprintf("select title, body, notifications_count, notification_info from notification where user_id = ? "+
		"and event_type in (%v)", strings.Join(templates, ","),
	), userId).WithContext(ctx).PageSize(10).PageState(pageState).Iter()

	nextPageState := iter.PageState()
	scanner := iter.Scanner()

	notifications := make([]database.Notification, 0)
	notificationsCounts := make(map[uuid.UUID]int64)

	for scanner.Next() {
		var title string
		var body string
		var notificationsCount int64
		var notificationInfo string

		if err := scanner.Scan(&title, &body, &notificationsCount, &notificationInfo); err != nil {
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

	nextPage := base64.StdEncoding.EncodeToString(nextPageState)

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

//func getFrontendSupportedNotificationTypes() []string { // temp fix https://tracki-workspace.slack.com/archives/C02LP6X90PL/p1648825668159999?thread_ts=1648825150.159869&cid=C02LP6X90PL
//	return []string{
//		"push.content.comment",
//		"push.profile.comment",
//		"push.comment.reply",
//		"push.profile.following",
//		"system",
//		"push.admin.bulk",
//		"push.comment.vote",
//		"push.content.like",
//		"push.bonus.daily",
//		"push.bonus.followers",
//		"push.content.successful-upload",
//		"push.content.new-posted",
//		"push.tip",
//		"push.content.rejected",
//		"push.kyc.status",
//		"push.content-creator.status",
//		"push.referral.other",
//		"push.referral.first",
//		"push.referral.megabonus",
//	}
//}

func getNotificationsTypesByTypeGroup(typeGroup TypeGroup) []string {
	switch typeGroup {
	case TypeGroupAll:
		var all = []string{"push.comment.new", "push.comment.reply", "push.comment.vote", "push.profile.comment",
			"push.content.comment", "push.admin.bulk", "push.admin.single", "push.profile.following",
			"push.content.new-posted", "push.like.new", "push.tip", "push.content.like", "push.bonus.followers",
			"push.bonus.daily", "push.content.successful-upload", "push.spot.successful-upload", "push.content.rejected", "push.kyc.status",
			"push.content-creator.status"}
		all = append(all, database.GetMarketingNotifications()...)
		return all
	case TypeGroupComment:
		return []string{"push.comment.new", "push.comment.reply", "push.comment.vote", "push.profile.comment",
			"push.content.comment"}
	case TypeGroupSystem:
		return []string{"push.admin.bulk", "push.admin.single"}
	case TypeGroupFollowing:
		return []string{"push.profile.following"}
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

	if err := db.Offset(req.Offset).Limit(req.Limit).Find(&notifications).Error; err != nil {
		return nil, err
	}

	notificationsResp := mapNotificationsToResponseItems(notifications, nil, userGoWrapper, followWrapper, ctx)

	return &ListNotificationsByAdminResponse{
		Items:      notificationsResp,
		TotalCount: totalCount,
	}, nil
}
