package notification

import (
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/wrappers/follow"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/utils"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.elastic.co/apm"
	"sort"
)

func mapNotificationsToResponseItems(notifications []database.Notification, userGoWrapper user_go.IUserGoWrapper,
	followWrapper follow.IFollowWrapper, apmTransaction *apm.Transaction) []NotificationsResponseItem {
	mapped := make(map[uuid.UUID]*NotificationsResponseItem, len(notifications))
	relatedUsersIdsMap := map[int64]bool{}

	for _, notification := range notifications {
		mappedItem := &NotificationsResponseItem{
			Id:                   notification.Id,
			UserId:               notification.UserId,
			Type:                 notification.Type,
			Title:                notification.Title,
			Message:              notification.Message,
			RelatedUserId:        notification.RelatedUserId,
			CommentId:            notification.CommentId,
			Comment:              notification.Comment,
			ContentId:            notification.ContentId,
			QuestionId:           notification.QuestionId,
			KycStatus:            notification.KycStatus,
			ContentCreatorStatus: notification.ContentCreatorStatus,
			KycReason:            notification.KycReason,
			CreatedAt:            notification.CreatedAt,
		}

		if mappedItem.RelatedUserId.Valid {
			mappedItem.RelatedUser = &NotificationsResponseUser{
				Id: mappedItem.RelatedUserId.Int64,
			}
		}

		if mappedItem.ContentId.ValueOrZero() > 0 {
			mappedItem.Content = &NotificationsResponseContent{
				NotificationContent: *notification.Content,
				ThumbUrl:            utils.GetThumbnailUrl(notification.Content.VideoId),
			}
		}

		mapped[notification.Id] = mappedItem
	}

	relatedUsersIds := make([]int64, len(relatedUsersIdsMap))
	i := 0
	for userId := range relatedUsersIdsMap {
		relatedUsersIds[i] = userId
		i++
	}

	routines := []chan error{
		fillUsers(mapped, userGoWrapper, apmTransaction),
		fillUserBlock(mapped, userGoWrapper, apmTransaction),
		fillFollowData(mapped, followWrapper, apmTransaction),
	}

	for _, c := range routines {
		if err := <-c; err != nil {
			apm_helper.CaptureApmError(err, apmTransaction)
		}
	}

	finalResp := make([]NotificationsResponseItem, len(mapped))
	i = 0
	for _, notification := range mapped {
		finalResp[i] = *notification
		i++
	}

	sort.Slice(finalResp, func(i, j int) bool {
		return finalResp[i].CreatedAt.Unix() > finalResp[j].CreatedAt.Unix()
	})

	return finalResp
}

func fillUsers(notifications map[uuid.UUID]*NotificationsResponseItem, userGoWrapper user_go.IUserGoWrapper, apmTransaction *apm.Transaction) chan error {
	ch := make(chan error, 2)

	if len(notifications) == 0 {
		close(ch)

		return ch
	}

	go func() {
		defer func() {
			close(ch)
		}()

		var userIds []int64

		for _, notification := range notifications {
			if !notification.RelatedUserId.Valid {
				continue
			}

			hasUserId := false

			for _, userId := range userIds {
				if userId == notification.RelatedUserId.Int64 {
					hasUserId = true
					break
				}
			}

			if hasUserId {
				continue
			}

			userIds = append(userIds, notification.RelatedUserId.Int64)
		}

		resp := <-userGoWrapper.GetUsersDetails(userIds, apmTransaction, false)

		if resp.Error != nil {
			ch <- errors.Wrap(errors.New(resp.Error.Message), "fill users failed")
		}

		for _, notification := range notifications {
			userResp, hasUser := resp.Items[notification.RelatedUserId.Int64]

			if !hasUser {
				continue
			}

			firstName, lastName := userResp.GetFirstAndLastNameWithPrivacy()

			notification.RelatedUser.Username = userResp.Username
			notification.RelatedUser.Firstname = firstName
			notification.RelatedUser.Lastname = lastName
			notification.RelatedUser.Verified = true
			notification.RelatedUser.AvatarUrl = userResp.Avatar
			notification.RelatedUser.NamePrivacyStatus = userResp.NamePrivacyStatus
		}
	}()

	return ch
}

func fillUserBlock(notifications map[uuid.UUID]*NotificationsResponseItem, userBlockWrapper user_go.IUserGoWrapper,
	apmTransaction *apm.Transaction) chan error {
	ch := make(chan error, 2)

	if len(notifications) == 0 {
		close(ch)

		return ch
	}

	go func() {
		defer func() {
			close(ch)
		}()

		userBlockMap := map[int64]map[int64]bool{}
		for _, notification := range notifications {
			if !notification.RelatedUserId.Valid {
				continue
			}

			if _, ok := userBlockMap[notification.UserId]; ok {
				userBlockMap[notification.UserId][notification.RelatedUserId.Int64] = false
			} else {
				userBlockMap[notification.UserId] = map[int64]bool{notification.RelatedUserId.Int64: false}
			}
		}

		for userId, relatedUserMap := range userBlockMap {
			for relatedUserId := range relatedUserMap {
				resp := <-userBlockWrapper.GetUserBlock(relatedUserId, userId, apmTransaction, false)
				if resp.Error != nil {
					ch <- errors.Wrap(errors.New(resp.Error.Message), "fill user block failed")
					return
				}

				if resp.Data.Type != nil && *resp.Data.Type == user_go.BlockedUser {
					userBlockMap[userId][relatedUserId] = resp.Data.IsBlocked
				}
			}
		}

		for _, notification := range notifications {
			if !notification.RelatedUserId.Valid {
				continue
			}

			isBlocked, hasUser := userBlockMap[notification.UserId][notification.RelatedUserId.Int64]

			if !hasUser {
				continue
			}

			notification.RelatedUser.IsBlocked = isBlocked
		}
	}()

	return ch
}

func fillFollowData(notifications map[uuid.UUID]*NotificationsResponseItem, followWrapper follow.IFollowWrapper,
	apmTransaction *apm.Transaction) chan error {
	ch := make(chan error, 2)

	if len(notifications) == 0 {
		close(ch)

		return ch
	}

	go func() {
		defer func() {
			close(ch)
		}()

		userMap := map[int64]map[int64]*follow.RelationData{}

		for _, notification := range notifications {
			if !notification.RelatedUserId.Valid {
				continue
			}

			if _, ok := userMap[notification.RelatedUserId.Int64]; ok {
				userMap[notification.RelatedUserId.Int64][notification.UserId] = nil
			} else {
				userMap[notification.RelatedUserId.Int64] = map[int64]*follow.RelationData{notification.UserId: nil}
			}
		}

		for relatedUserId, valueMap := range userMap {
			ids := make([]int64, len(valueMap))
			i := 0

			for userId := range valueMap {
				ids[i] = userId
				i++
			}

			resp := <-followWrapper.GetUserFollowingRelationBulk(relatedUserId, ids, apmTransaction, false)
			if resp.Error != nil {
				ch <- errors.Wrap(errors.New(resp.Error.Message), "fill follow data failed")
				return
			}

			for userId, followData := range resp.Data {
				userMap[relatedUserId][userId] = &followData
			}
		}

		for _, notification := range notifications {
			if _, ok := userMap[notification.RelatedUserId.Int64]; !ok {
				continue
			}

			if resp, ok := userMap[notification.RelatedUserId.Int64][notification.UserId]; ok && resp != nil {
				notification.RelatedUser.IsFollower = resp.IsFollowing
				notification.RelatedUser.IsFollowing = resp.IsFollower
			}
		}
	}()

	return ch
}
