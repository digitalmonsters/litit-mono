package ad_moderation

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/ads-manager/configs"
	"github.com/digitalmonsters/ads-manager/pkg/common"
	"github.com/digitalmonsters/ads-manager/pkg/converter"
	"github.com/digitalmonsters/ads-manager/pkg/database"
	"github.com/digitalmonsters/go-common/callback"
	"github.com/digitalmonsters/go-common/wrappers/notification_handler"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IService interface {
	GetAdModerationRequests(req GetAdModerationRequest, tx *gorm.DB, ctx context.Context) (*GetAdModerationResponse, error)
	SetAdRejectReason(req SetAdRejectReasonRequest, tx *gorm.DB, ctx context.Context) (*common.AddModerationItem, []callback.Callback, error)
}

type service struct {
	notificationHandler notification_handler.INotificationHandlerWrapper
	converter           *converter.Converter
}

func (s service) GetAdModerationRequests(req GetAdModerationRequest, tx *gorm.DB, ctx context.Context) (*GetAdModerationResponse, error) {
	var items []database.AdCampaign
	var q = tx.Model(items)

	if req.UserId.Valid {
		q = q.Where("user_id = ?", req.UserId.ValueOrZero())
	}
	if len(req.Status) > 0 {
		q = q.Where("status in ?", req.Status)
	}
	if req.StartedAtFrom.Valid {
		q = q.Where("started_at >= ?", req.StartedAtFrom.ValueOrZero())
	}
	if req.StartedAtTo.Valid {
		q = q.Where("started_at <= ?", req.StartedAtTo.ValueOrZero())
	}
	if req.EndedAtFrom.Valid {
		q = q.Where("ended_at >= ?", req.EndedAtFrom.ValueOrZero())
	}
	if req.EndedAtTo.Valid {
		q = q.Where("ended_at <= ?", req.EndedAtTo.ValueOrZero())
	}
	if req.MaxThresholdExceeded.Valid {
		q = q.Where("(created_at <= NOW() - INTERVAL ? and status = ?)",
			gorm.Expr(fmt.Sprintf("'%v HOURS'", configs.GetAppConfig().ADS_MODERATION_SLA)), database.AdCampaignStatusPending)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return nil, err
	}
	if err := q.Order("created_at desc").Find(&items).Error; err != nil {
		return nil, err
	}
	models := s.converter.MapFromDbAddCampaign(items, ctx)
	return &GetAdModerationResponse{
		TotalCount: count,
		Items:      models,
	}, nil
}

func (s service) SetAdRejectReason(req SetAdRejectReasonRequest, tx *gorm.DB, ctx context.Context) (*common.AddModerationItem, []callback.Callback, error) {
	var campaign database.AdCampaign

	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", req.Id).First(&campaign).Error; err != nil {
		return nil, nil, errors.WithStack(err)
	}
	if campaign.Status != database.AdCampaignStatusPending {
		return nil, nil, errors.WithStack(errors.New("ad already moderated"))
	}
	campaign.RejectReasonId = req.RejectReasonId
	campaign.Status = req.Status

	if err := tx.Model(&campaign).Updates(map[string]interface{}{
		"reject_reason_id": req.RejectReasonId,
		"status":           req.Status,
	}).Error; err != nil {
		return nil, nil, errors.WithStack(err)
	}
	var callbacks []callback.Callback
	if req.Status == database.AdCampaignStatusModerated {
		callbacks = append(callbacks, func(ctx context.Context) error {
			if s.notificationHandler == nil {
				return nil
			}

			return (<-s.notificationHandler.EnqueueNotificationWithTemplate("ads_campaign_approved", campaign.UserId,
				map[string]string{}, nil, ctx)).Error
		})
	} else if req.Status == database.AdCampaignStatusReject {
		callbacks = append(callbacks, func(ctx context.Context) error {
			if s.notificationHandler == nil {
				return nil
			}

			return (<-s.notificationHandler.EnqueueNotificationWithTemplate("ads_campaign_rejected", campaign.UserId,
				map[string]string{}, nil, ctx)).Error
		})
	} else {
		return nil, nil, errors.WithStack(errors.New("invalid status"))
	}

	models := s.converter.MapFromDbAddCampaign([]database.AdCampaign{campaign}, ctx)

	return &models[0], callbacks, nil
}

func NewService(notificationHandler notification_handler.INotificationHandlerWrapper, converter *converter.Converter) IService {
	return &service{
		notificationHandler: notificationHandler,
		converter:           converter,
	}
}
