package notification

import (
	"github.com/digitalmonsters/go-common/wrappers/follow"
	"github.com/digitalmonsters/go-common/wrappers/user_block"
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
	userBlockWrapper user_block.IUserBlockWrapper, followWrapper follow.IFollowWrapper, apmTransaction *apm.Transaction) (*NotificationsResponse, error) {
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

	notificationsResp := mapNotificationsToResponseItems(notifications, userGoWrapper, userBlockWrapper, followWrapper, apmTransaction)

	return &NotificationsResponse{
		Data:        notificationsResp,
		Next:        *cursor.After,
		Prev:        *cursor.Before,
		UnreadCount: userNotification.UnreadCount,
	}, nil
}

func getNotificationsTypesByTypeGroup(typeGroup TypeGroup) []string {
	switch typeGroup {
	case TypeGroupAll:
		return []string{"push.comment.new", "push.comment.reply", "push.comment.vote", "push.profile.comment",
			"push.content.comment", "push.admin.bulk", "push.admin.single", "push.profile.following",
			"push.content.new-posted", "push.like.new", "push.tip", "push.content.like", "push.bonus.followers",
			"push.bonus.daily", "push.content.successful-upload", "push.content.rejected", "push.kyc.status",
			"push.content-creator.status"}
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
	if err := db.Exec("insert into user_notifications(user_id, unread_count) values(?, 1) on conflict (user_id) do update set unread_count = unread_count + 1", userId).Error; err != nil {
		return err
	}

	return nil
}

func ListNotificationsByAdmin(db *gorm.DB, req ListNotificationsByAdminRequest, userGoWrapper user_go.IUserGoWrapper,
	userBlockWrapper user_block.IUserBlockWrapper, followWrapper follow.IFollowWrapper,
	apmTransaction *apm.Transaction) (*ListNotificationsByAdminResponse, error) {
	notifications := make([]database.Notification, 0)

	var userNotification database.UserNotification

	query := db.Model(userNotification)

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

	if err := db.Offset(req.Offset).Limit(req.Limit).Find(&userNotification).Error; err != nil {
		return nil, err
	}

	notificationsResp := mapNotificationsToResponseItems(notifications, userGoWrapper, userBlockWrapper, followWrapper, apmTransaction)

	return &ListNotificationsByAdminResponse{
		Items:      notificationsResp,
		TotalCount: totalCount,
	}, nil
}
