package notification

import (
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
	followWrapper follow.IFollowWrapper, apmTransaction *apm.Transaction) (*NotificationsResponse, error) {
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
		Where("user_id = ? and type in ?", userId, getNotificationsTypesByTypeGroup(typeGroup)).
		Where("type in ?", getFrontendSupportedNotificationTypes())
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

	notificationsResp := mapNotificationsToResponseItems(notifications, userGoWrapper, followWrapper, apmTransaction)

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

func getFrontendSupportedNotificationTypes() []string { // temp fix https://tracki-workspace.slack.com/archives/C02LP6X90PL/p1648825668159999?thread_ts=1648825150.159869&cid=C02LP6X90PL
	return []string{
		"push.content.comment",
		"push.profile.comment",
		"push.comment.reply",
		"push.profile.following",
		"system",
		"push.admin.bulk",
		"push.comment.vote",
		"push.content.like",
		"push.bonus.daily",
		"push.bonus.followers",
		"push.content.successful-upload",
		"push.content.new-posted",
		"push.tip",
		"push.content.rejected",
		"push.kyc.status",
		"push.content-creator.status",
		"push.referral.other",
		"push.referral.first",
		"push.referral.megabonus",
	}
}

func getNotificationsTypesByTypeGroup(typeGroup TypeGroup) []string {
	switch typeGroup {
	case TypeGroupAll:
		var all = []string{"push.comment.new", "push.comment.reply", "push.comment.vote", "push.profile.comment",
			"push.content.comment", "push.admin.bulk", "push.admin.single", "push.profile.following",
			"push.content.new-posted", "push.like.new", "push.tip", "push.content.like", "push.bonus.followers",
			"push.bonus.daily", "push.content.successful-upload", "push.content.rejected", "push.kyc.status",
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
	followWrapper follow.IFollowWrapper, apmTransaction *apm.Transaction) (*ListNotificationsByAdminResponse, error) {
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

	notificationsResp := mapNotificationsToResponseItems(notifications, userGoWrapper, followWrapper, apmTransaction)

	return &ListNotificationsByAdminResponse{
		Items:      notificationsResp,
		TotalCount: totalCount,
	}, nil
}
