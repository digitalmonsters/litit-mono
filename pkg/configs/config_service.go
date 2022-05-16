package configs

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/configurator/pkg/database"
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/callback"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"github.com/thoas/go-funk"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"strconv"
	"strings"
)

type IConfigService interface {
	GetAllConfigs(db *gorm.DB) ([]database.Config, error)
	GetConfigsByIds(db *gorm.DB, ids []string) ([]database.Config, error)
	AdminGetConfigs(db *gorm.DB, req GetConfigRequest) (*GetConfigResponse, error)
	AdminUpsertConfig(db *gorm.DB, req UpsertConfigRequest, userId int64, publisher eventsourcing.Publisher[eventsourcing.ConfigEvent]) (*application.ConfigModel, []callback.Callback, error)
	AdminGetConfigLogs(db *gorm.DB, req GetConfigLogsRequest) (*GetConfigLogsResponse, error)
	MigrateConfigs(db *gorm.DB, newConfigs map[string]application.MigrateConfigModel, publisher eventsourcing.Publisher[eventsourcing.ConfigEvent]) ([]application.ConfigModel, []callback.Callback, error)
}

type ConfigService struct {
}

var allConfigTypes = []application.ConfigType{application.ConfigTypeDecimal, application.ConfigTypeInteger, application.ConfigTypeBool,
	application.ConfigTypeString, application.ConfigTypeObject}
var allCategoryTypes = []application.ConfigCategory{application.ConfigCategoryAd, application.ConfigCategoryApplications,
	application.ConfigCategoryTokens, application.ConfigCategoryContent}

func (c *ConfigService) GetAllConfigs(db *gorm.DB) ([]database.Config, error) {
	var cfg []database.Config

	if err := db.Find(&cfg).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return cfg, nil
}

func (c *ConfigService) GetConfigsByIds(db *gorm.DB, ids []string) ([]database.Config, error) {
	var cfg []database.Config

	if err := db.Where("key in ?", ids).Find(&cfg).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return cfg, nil
}

func (c *ConfigService) AdminGetConfigs(db *gorm.DB, req GetConfigRequest) (*GetConfigResponse, error) {
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
		var search = fmt.Sprintf("%%%s%%", req.DescriptionContains.ValueOrZero())
		q = q.Where("(description ilike ? or key ilike ?)", search, search)
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
	if len(req.ReleaseVersions) > 0 {
		q = q.Where("release_version in ?", req.ReleaseVersions)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return nil, err
	}
	if err := q.Order("created_at desc").Find(&cfg).Error; err != nil {
		return nil, err
	}
	var respItems []application.ConfigModel
	for _, c := range cfg {
		respItems = append(respItems, application.ConfigModel{
			Key:            c.Key,
			Value:          c.Value,
			Type:           c.Type,
			Description:    c.Description,
			AdminOnly:      c.AdminOnly,
			CreatedAt:      c.CreatedAt,
			UpdatedAt:      c.UpdatedAt,
			Category:       c.Category,
			ReleaseVersion: c.ReleaseVersion,
		})
	}
	return &GetConfigResponse{
		Items:      respItems,
		TotalCount: count,
	}, nil
}

func validateNewConfigRequest(req UpsertConfigRequest) error {
	if len(req.Key) == 0 {
		return errors.New("invalid key")
	}
	if len(req.Type) == 0 {
		return errors.New("invalid type")
	} else if !funk.Contains(allConfigTypes, req.Type) {
		return errors.New("type does not exist")
	}
	if len(req.Value) == 0 {
		return errors.New("invalid value")
	}
	if len(req.Category) == 0 {
		return errors.New("invalid category")
	} else if !funk.Contains(allCategoryTypes, req.Category) {
		return errors.New("category does not exist")
	}
	if len(req.Description) == 0 {
		return errors.New("invalid description")
	}
	if len(req.ReleaseVersion) == 0 {
		return errors.New("invalid release version")
	}
	return validateValueAccordingToType(req.Type, req.Value)
}
func validateValueAccordingToType(reqType application.ConfigType, value string) error {
	var err error
	switch reqType {
	case application.ConfigTypeInteger:
		_, err = strconv.Atoi(value)
	case application.ConfigTypeDecimal:
		_, err = decimal.NewFromString(value)
	case application.ConfigTypeBool:
		var lowerVal = strings.ToLower(value)
		if lowerVal != "true" && lowerVal != "false" {
			return errors.New("invalid value")
		}
	}
	return err
}

func (c ConfigService) AdminUpsertConfig(tx *gorm.DB, req UpsertConfigRequest, userId int64, publisher eventsourcing.Publisher[eventsourcing.ConfigEvent]) (*application.ConfigModel, []callback.Callback, error) {
	if err := validateNewConfigRequest(req); err != nil {
		return nil, nil, err
	}
	var currentConfig database.Config
	if err := tx.Where("key = ?", req.Key).Find(&currentConfig).Error; err != nil {
		return nil, nil, err
	}
	if len(currentConfig.Key) > 0 {
		if len(currentConfig.Type) == 0 {
			currentConfig.Type = req.Type
		}
		if len(currentConfig.Category) == 0 {
			currentConfig.Category = req.Category
		}
		if len(currentConfig.ReleaseVersion) == 0 {
			currentConfig.ReleaseVersion = req.ReleaseVersion
		}
		currentConfig.Description = req.Description
		currentConfig.Value = req.Value

		if err := tx.Where("key = ?", currentConfig.Key).Save(&currentConfig).Error; err != nil {
			return nil, nil, err
		}
		if err := tx.Create(&database.ConfigLog{
			Key:           req.Key,
			Value:         req.Value,
			RelatedUserId: null.IntFrom(userId),
		}).Error; err != nil {
			return nil, nil, err
		}
	} else {
		return nil, nil, errors.New("config doesn't exist")
	}
	callbacks := []callback.Callback{
		func(ctx context.Context) error {
			if publisher != nil {
				err := <-publisher.PublishImmediate(ctx, eventsourcing.ConfigEvent{
					Key:   currentConfig.Key,
					Value: currentConfig.Value,
				})
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
	return &application.ConfigModel{
		Key:            currentConfig.Key,
		Value:          currentConfig.Value,
		Type:           currentConfig.Type,
		Description:    currentConfig.Description,
		AdminOnly:      currentConfig.AdminOnly,
		CreatedAt:      currentConfig.CreatedAt,
		UpdatedAt:      currentConfig.UpdatedAt,
		Category:       currentConfig.Category,
		ReleaseVersion: currentConfig.ReleaseVersion,
	}, callbacks, nil
}

func (c *ConfigService) AdminGetConfigLogs(db *gorm.DB, req GetConfigLogsRequest) (*GetConfigLogsResponse, error) {
	var items []database.ConfigLog

	var q = db.Model(items)
	if len(req.Keys) > 0 {
		q = q.Where("key in ?", req.Keys)
	}
	if req.KeyContains.Valid {
		q = q.Where("key ilike ?", fmt.Sprintf("%%%s%%", req.KeyContains.ValueOrZero()))
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

func (c *ConfigService) MigrateConfigs(tx *gorm.DB, newConfigs map[string]application.MigrateConfigModel,
	publisher eventsourcing.Publisher[eventsourcing.ConfigEvent]) ([]application.ConfigModel, []callback.Callback, error) {
	var keys []string

	for key := range newConfigs {
		keys = append(keys, key)
	}
	var currentConfigsKey []string
	if err := tx.Model(database.Config{}).Where("key in ?", keys).Pluck("key", &currentConfigsKey).Error; err != nil {
		return nil, nil, err
	}
	if len(currentConfigsKey) == len(keys) {
		return []application.ConfigModel{}, nil, nil
	}
	var currentConfigsMap = make(map[string]bool)

	for _, val := range currentConfigsKey {
		currentConfigsMap[val] = false
	}
	var events []eventsourcing.ConfigEvent
	var configModels []application.ConfigModel

	for key, val := range newConfigs {
		if _, ok := currentConfigsMap[key]; !ok {
			var newConfig = database.Config{
				Key:            val.Key,
				Value:          val.Value,
				Type:           val.Type,
				Description:    val.Description,
				AdminOnly:      val.AdminOnly,
				Category:       val.Category,
				ReleaseVersion: val.ReleaseVersion,
			}
			if err := tx.Create(&newConfig).Error; err != nil {
				return nil, nil, err
			}
			if err := tx.Create(&database.ConfigLog{
				Key:   val.Key,
				Value: val.Value,
			}).Error; err != nil {
				return nil, nil, err
			}
			events = append(events, eventsourcing.ConfigEvent{Key: val.Key, Value: val.Value})
			configModels = append(configModels, application.ConfigModel{
				Key:            newConfig.Key,
				Value:          newConfig.Value,
				Type:           newConfig.Type,
				Description:    newConfig.Description,
				AdminOnly:      newConfig.AdminOnly,
				Category:       newConfig.Category,
				ReleaseVersion: newConfig.ReleaseVersion,
				CreatedAt:      newConfig.CreatedAt,
				UpdatedAt:      newConfig.UpdatedAt,
			})
		}
	}

	callbacks := []callback.Callback{
		func(ctx context.Context) error {
			if publisher != nil {
				err := <-publisher.PublishImmediate(ctx, events...)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
	return configModels, callbacks, nil
}
