package configs

import (
	"github.com/digitalmonsters/configurator/pkg/database"
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/callback"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/router"
	"gorm.io/gorm"
)

type ConfigServiceMock struct {
	GetAllConfigsFn      func(db *gorm.DB) ([]database.Config, error)
	GetConfigsByIdsFn    func(db *gorm.DB, ids []string) ([]database.Config, error)
	AdminGetConfigsFn    func(db *gorm.DB, req GetConfigRequest, executionData router.MethodExecutionData) (*GetConfigResponse, error)
	AdminUpsertConfigFn  func(db *gorm.DB, req UpsertConfigRequest, userId int64, publisher eventsourcing.Publisher[eventsourcing.ConfigEvent]) (*application.ConfigModel, []callback.Callback, error)
	AdminGetConfigLogsFn func(db *gorm.DB, req GetConfigLogsRequest, executionData router.MethodExecutionData) (*GetConfigLogsResponse, error)
	MigrateConfigsFn     func(db *gorm.DB, newConfigs map[string]application.MigrateConfigModel, publisher eventsourcing.Publisher[eventsourcing.ConfigEvent]) ([]application.ConfigModel, []callback.Callback, error)
}

func (c *ConfigServiceMock) GetAllConfigs(db *gorm.DB) ([]database.Config, error) {
	return c.GetAllConfigsFn(db)
}
func (c *ConfigServiceMock) GetConfigsByIds(db *gorm.DB, ids []string) ([]database.Config, error) {
	return c.GetConfigsByIdsFn(db, ids)
}
func (c *ConfigServiceMock) AdminGetConfigs(db *gorm.DB, req GetConfigRequest, executionData router.MethodExecutionData) (*GetConfigResponse, error) {
	return c.AdminGetConfigsFn(db, req, executionData)
}
func (c *ConfigServiceMock) AdminUpsertConfig(db *gorm.DB, req UpsertConfigRequest, userId int64,
	publisher eventsourcing.Publisher[eventsourcing.ConfigEvent]) (*application.ConfigModel, []callback.Callback, error) {
	return c.AdminUpsertConfigFn(db, req, userId, publisher)
}
func (c *ConfigServiceMock) AdminGetConfigLogs(db *gorm.DB, req GetConfigLogsRequest, executionData router.MethodExecutionData) (*GetConfigLogsResponse, error) {
	return c.AdminGetConfigLogsFn(db, req, executionData)
}
func (c *ConfigServiceMock) MigrateConfigs(db *gorm.DB, newConfigs map[string]application.MigrateConfigModel, publisher eventsourcing.Publisher[eventsourcing.ConfigEvent]) ([]application.ConfigModel, []callback.Callback, error) {
	return c.MigrateConfigsFn(db, newConfigs, publisher)
}

func GetMock() IConfigService {
	return &ConfigServiceMock{}
}
