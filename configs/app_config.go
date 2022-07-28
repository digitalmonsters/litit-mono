package configs

import (
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
	ADS_MODERATION_SLA int
}

func GetConfigsMigration() map[string]application.MigrateConfigModel {
	return map[string]application.MigrateConfigModel{
		"ADS_MODERATION_SLA": {
			Key:            "ADS_MODERATION_SLA",
			Value:          "48", //hours
			Type:           application.ConfigTypeInteger,
			Description:    "Ads moderation SLA",
			Category:       application.ConfigCategoryAd,
			ReleaseVersion: "29.07.22",
		},
	}
}
