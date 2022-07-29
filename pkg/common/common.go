package common

import (
	"github.com/digitalmonsters/ads-manager/pkg/database"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type IService interface {
	ListActionButtons(request ListActionButtonsRequest, tx *gorm.DB) (*ListActionButtonsResponse, error)
	UpsertActionButtons(req UpsertActionButtonsRequest, tx *gorm.DB) error
	DeleteActionButtons(req DeleteRequest, tx *gorm.DB) error
	ListRejectReasons(request ListRejectReasonsRequest, tx *gorm.DB) (*ListRejectReasonsResponse, error)
	UpsertRejectReasons(req UpsertRejectReasonsRequest, tx *gorm.DB) error
	DeleteRejectReasons(req DeleteRequest, tx *gorm.DB) error
}

type service struct {
}

func (s service) ListActionButtons(request ListActionButtonsRequest, tx *gorm.DB) (*ListActionButtonsResponse, error) {
	var count int64
	var items []database.ActionButton

	if err := tx.Model(items).Count(&count).Error; err != nil {
		return nil, err
	}
	if err := tx.Limit(request.Limit).Offset(request.Offset).Find(&items).Error; err != nil {
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

func (s service) ListRejectReasons(request ListRejectReasonsRequest, tx *gorm.DB) (*ListRejectReasonsResponse, error) {
	var count int64
	var items []database.RejectReason

	if err := tx.Model(items).Count(&count).Error; err != nil {
		return nil, err
	}
	if err := tx.Limit(request.Limit).Offset(request.Offset).Find(&items).Error; err != nil {
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

func (s service) UpsertActionButtons(req UpsertActionButtonsRequest, tx *gorm.DB) error {
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

func (s service) DeleteActionButtons(req DeleteRequest, tx *gorm.DB) error {
	var count int64

	if err := tx.Model(database.AdCampaign{}).Where("link_button_id in ?", req.Ids).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.WithStack(errors.New("action buttons in use"))
	}
	return tx.Where("id in ?", req.Ids).Delete(database.ActionButton{}).Error
}

func (s service) UpsertRejectReasons(req UpsertRejectReasonsRequest, tx *gorm.DB) error {
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

func (s service) DeleteRejectReasons(req DeleteRequest, tx *gorm.DB) error {
	var count int64

	if err := tx.Model(database.AdCampaign{}).Where("reject_reason_id in ?", req.Ids).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.WithStack(errors.New("reject reasons in use"))
	}
	return tx.Where("id in ?", req.Ids).Delete(database.RejectReason{}).Error
}

func NewService() IService {
	return &service{}
}
