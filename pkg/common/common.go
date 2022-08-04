package common

import (
	"github.com/digitalmonsters/ads-manager/pkg/database"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"time"
)

type IService interface {
	ListActionButtons(request ListActionButtonsRequest, tx *gorm.DB) (*ListActionButtonsResponse, error)
	UpsertActionButtons(req UpsertActionButtonsRequest, tx *gorm.DB) error
	DeleteActionButtons(req DeleteRequest, tx *gorm.DB) error
	ListRejectReasons(request ListRejectReasonsRequest, tx *gorm.DB) (*ListRejectReasonsResponse, error)
	UpsertRejectReasons(req UpsertRejectReasonsRequest, tx *gorm.DB) error
	DeleteRejectReasons(req DeleteRequest, tx *gorm.DB) error
	PublicListActionButtons(request PublicListActionButtonsRequest, tx *gorm.DB) (*ListActionButtonsResponse, error)
	UpsertAdCampaignCountryPrice(request UpsertAdCampaignCountryPriceRequest, tx *gorm.DB) error
	ListAdCampaignCountryPrices(request ListAdCampaignCountryPriceRequest, tx *gorm.DB) (*ListAdCampaignCountryPriceResponse, error)
}

type service struct {
}

func (s *service) ListActionButtons(request ListActionButtonsRequest, tx *gorm.DB) (*ListActionButtonsResponse, error) {
	var count int64
	var items []database.ActionButton

	var q = tx.Model(items).Where("deleted_at is null")

	if err := q.Count(&count).Error; err != nil {
		return nil, err
	}
	if err := q.Limit(request.Limit).Offset(request.Offset).Find(&items).Error; err != nil {
		return nil, err
	}

	var models []ActionButtonModel

	for _, item := range items {
		models = append(models, ActionButtonModel{
			Id:   item.Id,
			Type: item.Type,
			Name: item.Name,
		})
	}
	return &ListActionButtonsResponse{
		Items:      models,
		TotalCount: count,
	}, nil
}

func (s *service) ListRejectReasons(request ListRejectReasonsRequest, tx *gorm.DB) (*ListRejectReasonsResponse, error) {
	var count int64
	var items []database.RejectReason

	var q = tx.Model(items).Where("deleted_at is null")

	if err := q.Count(&count).Error; err != nil {
		return nil, err
	}
	if err := q.Limit(request.Limit).Offset(request.Offset).Find(&items).Error; err != nil {
		return nil, err
	}

	var models []RejectReasonModel

	for _, item := range items {
		models = append(models, RejectReasonModel{
			Id:     item.Id,
			Reason: item.Reason,
		})
	}
	return &ListRejectReasonsResponse{
		Items:      models,
		TotalCount: count,
	}, nil
}

func (s *service) UpsertActionButtons(req UpsertActionButtonsRequest, tx *gorm.DB) error {
	var creates []database.ActionButton
	for _, item := range req.Items {
		if item.Id.Valid {
			if err := tx.Model(database.ActionButton{}).Where("id = ?", item.Id.ValueOrZero()).Updates(map[string]interface{}{
				"type": item.Type,
				"name": item.Name,
			}).Error; err != nil {
				return err
			}
		} else {
			creates = append(creates, database.ActionButton{
				Name: item.Name,
				Type: item.Type,
			})
		}
	}
	if len(creates) > 0 {
		if err := tx.Create(&creates).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *service) DeleteActionButtons(req DeleteRequest, tx *gorm.DB) error {
	var count int64

	if err := tx.Model(database.AdCampaign{}).Where("link_button_id in ?", req.Ids).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.WithStack(errors.New("action buttons in use"))
	}
	return tx.Model(database.ActionButton{}).Where("id in ?", req.Ids).Update("deleted_at", time.Now().UTC()).Error
}

func (s *service) UpsertRejectReasons(req UpsertRejectReasonsRequest, tx *gorm.DB) error {
	var creates []database.RejectReason
	for _, item := range req.Items {
		if item.Id.Valid {
			if err := tx.Model(database.RejectReason{}).Where("id = ?", item.Id.ValueOrZero()).Updates(map[string]interface{}{
				"reason": item.Reason,
			}).Error; err != nil {
				return err
			}
		} else {
			creates = append(creates, database.RejectReason{
				Reason: item.Reason,
			})
		}
	}
	if len(creates) > 0 {
		if err := tx.Create(&creates).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *service) DeleteRejectReasons(req DeleteRequest, tx *gorm.DB) error {
	var count int64

	if err := tx.Model(database.AdCampaign{}).Where("reject_reason_id in ?", req.Ids).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.WithStack(errors.New("reject reasons in use"))
	}
	return tx.Model(database.RejectReason{}).Where("id in ?", req.Ids).Update("deleted_at", time.Now().UTC()).Error
}

func NewService() IService {
	return &service{}
}

func (s *service) PublicListActionButtons(request PublicListActionButtonsRequest, tx *gorm.DB) (*ListActionButtonsResponse, error) {
	var count int64
	var items []database.ActionButton

	var q = tx.Model(items).Where("deleted_at is null")

	if request.Type.Valid {
		q = q.Where("type = ?", request.Type.ValueOrZero())
	}
	if err := q.Count(&count).Error; err != nil {
		return nil, err
	}
	if request.Limit > 0 {
		q = q.Limit(request.Limit)
	}
	if err := q.Offset(request.Offset).Find(&items).Error; err != nil {
		return nil, err
	}
	var models []ActionButtonModel

	for _, item := range items {
		models = append(models, ActionButtonModel{
			Id:   item.Id,
			Type: item.Type,
			Name: item.Name,
		})
	}
	return &ListActionButtonsResponse{
		Items:      models,
		TotalCount: count,
	}, nil
}

func (s *service) UpsertAdCampaignCountryPrice(request UpsertAdCampaignCountryPriceRequest, tx *gorm.DB) error {
	return nil
}
func (s *service) ListAdCampaignCountryPrices(request ListAdCampaignCountryPriceRequest, tx *gorm.DB) (*ListAdCampaignCountryPriceResponse, error) {
	return nil, nil
}
