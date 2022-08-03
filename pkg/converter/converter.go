package converter

import (
	"context"
	"github.com/digitalmonsters/ads-manager/configs"
	"github.com/digitalmonsters/ads-manager/pkg/common"
	"github.com/digitalmonsters/ads-manager/pkg/database"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"time"
)

type Converter struct {
	userWrapper user_go.IUserGoWrapper
}

func NewConverter(userWrapper user_go.IUserGoWrapper) *Converter {
	return &Converter{
		userWrapper: userWrapper,
	}
}

func (c *Converter) MapFromDbAddCampaign(dbModels []database.AdCampaign, ctx context.Context) []common.AddModerationItem {
	var userIds []int64
	var items []*common.AddModerationItem

	for _, dbModel := range dbModels {
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
	routines := []chan error{
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
