package configs

import (
	"context"
	"github.com/digitalmonsters/configurator/pkg/database"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"gorm.io/gorm"
)

type ConfigServiceMock struct {
	GetAllConfigsFn      func(db *gorm.DB) ([]database.Config, error)
	GetConfigsByIdsFn    func(db *gorm.DB, ids []string) ([]database.Config, error)
	AdminGetConfigsFn    func(db *gorm.DB, req GetConfigRequest) (*GetConfigResponse, error)
	AdminUpsertConfigFn  func(db *gorm.DB, req UpsertConfigRequest, userId int64, publisher eventsourcing.Publisher[ConfigEvent], ctx context.Context) (*ConfigModel, error)
	AdminGetConfigLogsFn func(db *gorm.DB, req GetConfigLogsRequest) (*GetConfigLogsResponse, error)
}

func (c *ConfigServiceMock) GetAllConfigs(db *gorm.DB) ([]database.Config, error) {
	return c.GetAllConfigsFn(db)
}
func (c *ConfigServiceMock) GetConfigsByIds(db *gorm.DB, ids []string) ([]database.Config, error) {
	return c.GetConfigsByIdsFn(db, ids)
}
func (c *ConfigServiceMock) AdminGetConfigs(db *gorm.DB, req GetConfigRequest) (*GetConfigResponse, error) {
	return c.AdminGetConfigsFn(db, req)
}
func (c *ConfigServiceMock) AdminUpsertConfig(db *gorm.DB, req UpsertConfigRequest, userId int64,
	publisher eventsourcing.Publisher[ConfigEvent], ctx context.Context) (*ConfigModel, error) {
	return c.AdminUpsertConfigFn(db, req, userId, publisher, ctx)
}
func (c *ConfigServiceMock) AdminGetConfigLogs(db *gorm.DB, req GetConfigLogsRequest) (*GetConfigLogsResponse, error) {
	return c.AdminGetConfigLogsFn(db, req)
}

func GetMock() IConfigService {
	return &ConfigServiceMock{}
}
