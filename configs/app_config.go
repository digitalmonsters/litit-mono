package configs

import (
	"fmt"
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/boilerplate_testing"
)

var cfgService *application.Configurator[AppConfig]
var mockAppConfigs = map[string]AppConfig{}

func SetMockAppConfig(mock AppConfig) {
	boilerplate_testing.SetMockAppConfig(mockAppConfigs, mock)
}

func GetAppConfig() AppConfig {
	if boilerplate.GetCurrentEnvironment() == boilerplate.Ci || boilerplate.GetCurrentEnvironment() == boilerplate.Local {
		return boilerplate_testing.GetMockAppConfig(mockAppConfigs)
	}

	return cfgService.Values
}

type AppConfig struct {
	MUSIC_MAX_HASHTAGS_COUNT int
}

func GetConfigsMigration() map[string]application.MigrateConfigModel {
	return map[string]application.MigrateConfigModel{
		"MUSIC_MAX_HASHTAGS_COUNT": {
			Key:            "MUSIC_MAX_HASHTAGS_COUNT",
			Value:          fmt.Sprint(10),
			Type:           application.ConfigTypeInteger,
			Description:    "Max hashtags count in music",
			Category:       application.ConfigCategoryApplications,
			ReleaseVersion: "22.07.22",
		},
	}
}
