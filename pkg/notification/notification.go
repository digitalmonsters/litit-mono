package notification

import (
	"context"
	"github.com/digitalmonsters/go-common/wrappers/follow"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/google/uuid"
	"github.com/pilagod/gorm-cursor-paginator/v2/paginator"
	"github.com/pkg/errors"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
)

func GetNotifications(db *gorm.DB, userId int64, page string, typeGroup TypeGroup, userGoWrapper user_go.IUserGoWrapper,
	followWrapper follow.IFollowWrapper, apmTransaction *apm.Transaction, ctx context.Context) (*NotificationsResponse, error) {
	notifications := make([]database.Notification, 0)

	p := paginator.New(
		&paginator.Config{
			Rules: []paginator.Rule{{
				Key:   "CreatedAt",
				Order: paginator.DESC,
			}},
			Limit: 10,
			After: page,
		},
	)

	query := db.Model(notifications).
		Where("user_id = ? and type in ?", userId, getNotificationsTypesByTypeGroup(typeGroup))
	result, cursor, err := p.Paginate(query, &notifications)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if result.Error != nil {
		return nil, errors.WithStack(result.Error)
	}

	var userNotification database.UserNotification

	if err = db.Where("user_id = ?", userId).Find(&userNotification).Error; err != nil {
		return nil, err
	}

	notificationsResp := mapNotificationsToResponseItems(notifications, userGoWrapper, followWrapper, apmTransaction, ctx)

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
	followWrapper follow.IFollowWrapper, apmTransaction *apm.Transaction, ctx context.Context) (*ListNotificationsByAdminResponse, error) {
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

	notificationsResp := mapNotificationsToResponseItems(notifications, userGoWrapper, followWrapper, apmTransaction, ctx)

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
	session := database.GetScyllaSession()

	iter := session.Query("select notification_id, read_count from user_notifications_read_counter where notification_id in ?;",
		req.NotificationIds).
		WithContext(ctx).Iter()

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
