package configs

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/configurator/pkg/database"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type IConfigService interface {
	GetAllConfigs(db *gorm.DB) ([]database.Config, error)
	GetConfigsByIds(db *gorm.DB, ids []string) ([]database.Config, error)
	AdminGetConfigs(db *gorm.DB, req GetConfigRequest) (*GetConfigResponse, error)
	AdminUpsertConfig(db *gorm.DB, req UpsertConfigRequest, userId int64, publisher eventsourcing.Publisher[ConfigEvent], ctx context.Context) (*ConfigModel, error)
	AdminGetConfigLogs(db *gorm.DB, req GetConfigLogsRequest) (*GetConfigLogsResponse, error)
}

type ConfigService struct {
}

func (c ConfigService) GetAllConfigs(db *gorm.DB) ([]database.Config, error) {
	var cfg []database.Config

	if err := db.Find(&cfg).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return cfg, nil
}

func (c ConfigService) GetConfigsByIds(db *gorm.DB, ids []string) ([]database.Config, error) {
	var cfg []database.Config

	if err := db.Where("key in ?", ids).Find(&cfg).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return cfg, nil
}

func (c ConfigService) AdminGetConfigs(db *gorm.DB, req GetConfigRequest) (*GetConfigResponse, error) {
	var cfg []database.Config
	var q = db.Model(cfg)

	if req.AdminOnly.Valid {
		q = q.Where("admin_only = ?", req.AdminOnly)
	}
	if len(req.Categories) > 0 {
		q = q.Where("category in ?", req.Categories)
	}
	if len(req.Types) > 0 {
		q = q.Where("type in ?", req.Types)
	}
	if len(req.Keys) > 0 {
		q = q.Where("key in ?", req.Keys)
	}
	if req.DescriptionContains.Valid {
		q = q.Where("description ilike ?", fmt.Sprintf("%%%s%%", req.DescriptionContains.ValueOrZero()))
	}
	if req.CreatedFrom.Valid {
		q = q.Where("created_at > ?", req.CreatedFrom.Time)
	}
	if req.CreatedTo.Valid {
		q = q.Where("created_at < ?", req.CreatedTo.Time)
	}
	if req.UpdatedFrom.Valid {
		q = q.Where("updated_at > ?", req.CreatedFrom.Time)
	}
	if req.UpdatedTo.Valid {
		q = q.Where("updated_at < ?", req.CreatedTo.Time)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return nil, err
	}
	if err := q.Order("created_at desc").Find(&cfg).Error; err != nil {
		return nil, err
	}
	var respItems []ConfigModel
	for _, c := range cfg {
		respItems = append(respItems, ConfigModel{
			Key:         c.Key,
			Value:       c.Value,
			Type:        c.Type,
			Description: c.Description,
			AdminOnly:   c.AdminOnly,
			CreatedAt:   c.CreatedAt,
			UpdatedAt:   c.UpdatedAt,
			Category:    c.Category,
		})
	}
	return &GetConfigResponse{
		Items:      respItems,
		TotalCount: count,
	}, nil
}

func (c ConfigService) AdminUpsertConfig(db *gorm.DB, req UpsertConfigRequest, userId int64, publisher eventsourcing.Publisher[ConfigEvent], ctx context.Context) (*ConfigModel, error) {
	if len(req.Key) == 0 {
		return nil, errors.New("invalid key")
	}
	if len(req.Type) == 0 {
		return nil, errors.New("invalid type")
	}
	if len(req.Value) == 0 {
		return nil, errors.New("invalid value")
	}
	if len(req.Category) == 0 {
		return nil, errors.New("invalid category")
	}
	if len(req.Description) == 0 {
		return nil, errors.New("invalid decsription")
	}
	var currentConfig database.Config
	if err := db.Where("key = ?", req.Key).Find(&currentConfig).Error; err != nil {
		return nil, err
	}
	var tx = db.Begin()
	defer tx.Rollback()
	if len(currentConfig.Key) > 0 {
		if err := tx.Model(currentConfig).Where("key = ?", req.Key).Updates(map[string]interface{}{
			"value":       req.Value,
			"description": req.Description,
			"admin_only":  req.AdminOnly,
			"category":    req.Category,
			"type":        req.Type,
		}).Error; err != nil {
			return nil, err
		}
		if err := tx.Create(&database.ConfigLog{
			Key:           req.Key,
			Value:         req.Value,
			RelatedUserId: userId,
		}).Error; err != nil {
			return nil, err
		}
		if err := tx.Where("key = ?", req.Key).Find(&currentConfig).Error; err != nil {
			return nil, err
		}
	} else {
		currentConfig = database.Config{
			Key:         req.Key,
			Value:       req.Value,
			Type:        req.Type,
			Description: req.Description,
			AdminOnly:   req.AdminOnly,
			Category:    req.Category,
		}
		if err := tx.Create(&currentConfig).Error; err != nil {
			return nil, err
		}
		if err := tx.Create(&database.ConfigLog{
			Key:           req.Key,
			Value:         req.Value,
			RelatedUserId: userId,
		}).Error; err != nil {
			return nil, err
		}
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	if publisher != nil {
		err := <-publisher.PublishImmediate(ctx, ConfigEvent{
			Key:   currentConfig.Key,
			Value: currentConfig.Value,
		})
		apm_helper.LogError(err, ctx)
	}
	return &ConfigModel{
		Key:         currentConfig.Key,
		Value:       currentConfig.Value,
		Type:        currentConfig.Type,
		Description: currentConfig.Description,
		AdminOnly:   currentConfig.AdminOnly,
		CreatedAt:   currentConfig.CreatedAt,
		UpdatedAt:   currentConfig.UpdatedAt,
		Category:    currentConfig.Category,
	}, nil
}

func (c ConfigService) AdminGetConfigLogs(db *gorm.DB, req GetConfigLogsRequest) (*GetConfigLogsResponse, error) {
	var items []database.ConfigLog

	var q = db.Model(items)
	if len(req.Keys) > 0 {
		q = q.Where("key in ?", req.Keys)
	}
	if len(req.RelatedUserIds) > 0 {
		q = q.Where("related_user_id in ?", req.RelatedUserIds)
	}
	if req.CreatedFrom.Valid {
		q = q.Where("created_at > ?", req.CreatedFrom.Time)
	}
	if req.CreatedTo.Valid {
		q = q.Where("created_at < ?", req.CreatedTo.Time)
	}
	if req.UpdatedFrom.Valid {
		q = q.Where("updated_at > ?", req.CreatedFrom.Time)
	}
	if req.UpdatedTo.Valid {
		q = q.Where("updated_at < ?", req.CreatedTo.Time)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return nil, err
	}
	if err := q.Order("created_at desc").Find(&items).Error; err != nil {
		return nil, err
	}
	var respItems []ConfigLogModel
	for _, it := range items {
		respItems = append(respItems, ConfigLogModel{
			Id:            it.Id,
			Key:           it.Key,
			Value:         it.Value,
			CreatedAt:     it.CreatedAt,
			UpdatedAt:     it.UpdatedAt,
			RelatedUserId: it.RelatedUserId,
		})
	}
	return &GetConfigLogsResponse{
		Items:      respItems,
		TotalCount: count,
	}, nil
}
