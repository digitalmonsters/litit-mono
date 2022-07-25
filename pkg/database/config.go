package database

import (
	"github.com/digitalmonsters/go-common/application"
	"gopkg.in/guregu/null.v4"
	"time"
)

type Config struct {
	Key             string
	Value           string
	Type            application.ConfigType
	Description     string
	AdminOnly       bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Category        application.ConfigCategory
	ReleaseVersion  string
	LastChangedById null.Int
}

func (Config) TableName() string {
	return "configs"
}

type ConfigLog struct {
	Id            int64
	Key           string
	Value         string
	OldValue      string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	RelatedUserId null.Int
}

func (ConfigLog) TableName() string {
	return "config_logs"
}
