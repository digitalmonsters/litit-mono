package ad_campaign

import (
	"context"
	"github.com/digitalmonsters/ads-manager/pkg/database"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/pkg/errors"
	"go.elastic.co/apm"
	"gorm.io/gorm"
	"time"
)

type IService interface {
	CreateAdCampaign(req CreateAdCampaignRequest, userId int64, tx *gorm.DB, ctx context.Context) error
}

type service struct {
	contentWrapper content.IContentWrapper
}

func NewService(contentWrapper content.IContentWrapper) IService {
	return &service{
		contentWrapper: contentWrapper,
	}
}

func (s *service) CreateAdCampaign(req CreateAdCampaignRequest, userId int64, tx *gorm.DB, ctx context.Context) error {
	contentResp := <-s.contentWrapper.GetInternal([]int64{req.ContentId}, false, apm.TransactionFromContext(ctx), false)
	if contentResp.Error != nil {
		return errors.WithStack(contentResp.Error.ToError())
	}

	if _, ok := contentResp.Response[req.ContentId]; !ok {
		return errors.WithStack(errors.New("content not found"))
	}

	if err := tx.Create(&database.AdCampaign{
		UserId:       userId,
		Name:         req.Name,
		AdType:       req.AdType,
		Status:       database.AdCampaignStatusPending,
		ContentId:    req.ContentId,
		Link:         req.Link,
		LinkButtonId: req.LinkButtonId,
		Country:      req.Country,
		CreatedAt:    time.Now().UTC(),
		DurationMin:  req.DurationMin,
		Budget:       req.Budget,
		Gender:       req.Gender,
		AgeFrom:      req.AgeFrom,
		AgeTo:        req.AgeTo,
	}).Error; err != nil {
		return errors.WithStack(err)
	}

	return nil
}
