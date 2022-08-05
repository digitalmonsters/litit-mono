package converter

import (
	"context"
	"github.com/digitalmonsters/ads-manager/configs"
	"github.com/digitalmonsters/ads-manager/pkg/common"
	"github.com/digitalmonsters/ads-manager/pkg/database"
	"github.com/digitalmonsters/ads-manager/utils"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"go.elastic.co/apm"
	"time"
)

type Converter struct {
	userWrapper    user_go.IUserGoWrapper
	contentWrapper content.IContentWrapper
}

func NewConverter(userWrapper user_go.IUserGoWrapper, contentWrapper content.IContentWrapper) *Converter {
	return &Converter{
		userWrapper:    userWrapper,
		contentWrapper: contentWrapper,
	}
}

func (c *Converter) MapFromDbAddCampaign(dbModels []database.AdCampaign, ctx context.Context) []common.AddModerationItem {
	var userIds []int64
	var usersMap = make(map[int64]bool)

	var contentIds []int64

	var items []*common.AddModerationItem

	for _, dbModel := range dbModels {
		contentIds = append(contentIds, dbModel.ContentId)
		usersMap[dbModel.UserId] = false
		items = append(items, &common.AddModerationItem{
			Id:             dbModel.Id,
			UserId:         dbModel.UserId,
			Name:           dbModel.Name,
			AdType:         dbModel.AdType,
			Status:         dbModel.Status,
			ContentId:      dbModel.ContentId,
			Link:           dbModel.Link,
			Country:        dbModel.Country,
			CreatedAt:      dbModel.CreatedAt,
			StartedAt:      dbModel.StartedAt,
			EndedAt:        dbModel.EndedAt,
			DurationMin:    dbModel.DurationMin,
			Budget:         dbModel.Budget,
			OriginalBudget: dbModel.OriginalBudget,
			Gender:         dbModel.Gender,
			AgeFrom:        dbModel.AgeFrom,
			AgeTo:          dbModel.AgeTo,
			RejectReasonId: dbModel.RejectReasonId,
			SlaExpired: dbModel.Status == database.AdCampaignStatusPending &&
				time.Now().After(dbModel.CreatedAt.Add(time.Hour*time.Duration(configs.GetAppConfig().ADS_MODERATION_SLA))),
		})
	}

	for userId := range usersMap {
		userIds = append(userIds, userId)
	}

	routines := []chan error{
		c.fillContent(items, contentIds, ctx),
		c.fillUsers(items, userIds, ctx),
	}

	for _, c := range routines {
		if err := <-c; err != nil {
			apm_helper.LogError(err, ctx)
		}
	}
	var valItems []common.AddModerationItem
	for _, item := range items {
		valItems = append(valItems, *item)
	}
	return valItems
}

func (c *Converter) fillUsers(items []*common.AddModerationItem, userIds []int64, ctx context.Context) chan error {
	ch := make(chan error, 2)

	go func() {
		defer func() {
			close(ch)
		}()
		if len(userIds) == 0 {
			ch <- nil
		}
		userResp := <-c.userWrapper.GetUsers(userIds, ctx, false)
		if userResp.Error != nil {
			ch <- userResp.Error.ToError()
		}

		for _, item := range items {
			if val, ok := userResp.Response[item.UserId]; ok {
				item.Username = val.Username
				item.FirstName = val.Firstname
				item.LastName = val.Lastname
				item.Email = val.Email
			}
		}

	}()
	return ch
}

func (c *Converter) fillContent(items []*common.AddModerationItem, contentIds []int64, ctx context.Context) chan error {
	ch := make(chan error, 2)

	go func() {
		defer func() {
			close(ch)
		}()
		if len(contentIds) == 0 {
			ch <- nil
		}
		contentResp := <-c.contentWrapper.GetInternal(contentIds, false, apm.TransactionFromContext(ctx), false)
		if contentResp.Error != nil {
			ch <- contentResp.Error.ToError()
		}

		for _, item := range items {
			if val, ok := contentResp.Response[item.ContentId]; ok {
				item.VideoUrl = utils.GetVideoUrl(val.VideoId)
				item.Thumbnail = utils.GetThumbnailUrl(val.VideoId)
				item.AnimUrl = utils.GetAnimUrl(val.VideoId)
			}
		}

	}()
	return ch
}
