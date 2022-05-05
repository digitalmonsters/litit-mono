package database

import "time"

type Config struct {
	Key         string
	Value       string
	Type        ConfigType
	Description string
	AdminOnly   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Category    ConfigCategory
}

func (Config) TableName() string {
	return "configs"
}

type ConfigType string

const (
	ConfigTypeString = ConfigType("string")
	ConfigTypeNumber = ConfigType("number")
	ConfigTypeObject = ConfigType("object")
)

type ConfigCategory string

const (
	ConfigCategoryApplications = ConfigCategory("applications")
	ConfigCategoryTokens       = ConfigCategory("tokens")
	ConfigCategoryContent      = ConfigCategory("content")
	ConfigCategoryAd           = ConfigCategory("ad")
)

type ConfigLog struct {
	Id            int64
	Key           string
	Value         string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	RelatedUserId int64
}

func (ConfigLog) TableName() string {
	return "config_logs"
}
