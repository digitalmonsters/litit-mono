package configs

import (
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/shopspring/decimal"
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
	ADS_MODERATION_SLA                     int
	ADS_CAMPAIGN_VIDEOS_PER_CONTENT_VIDEOS int
	ADS_CAMPAIGN_GLOBAL_PRICE              decimal.Decimal
	ADS_AVAILABLE_FOR_USER_IDS             string
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
		"ADS_CAMPAIGN_VIDEOS_PER_CONTENT_VIDEOS": {
			Key:            "ADS_CAMPAIGN_VIDEOS_PER_CONTENT_VIDEOS",
			Value:          "9",
			Type:           application.ConfigTypeInteger,
			Description:    "Ads campaign videos per content videos frequency, possible values are: 2, 4, 5, 10, 20",
			Category:       application.ConfigCategoryAd,
			ReleaseVersion: "05.08.22",
		},
		"ADS_CAMPAIGN_GLOBAL_PRICE": {
			Key:            "ADS_CAMPAIGN_GLOBAL_PRICE",
			Value:          "10",
			Type:           application.ConfigTypeDecimal,
			Description:    "Ads campaign global price",
			Category:       application.ConfigCategoryAd,
			ReleaseVersion: "05.08.22",
		},
		"ADS_AVAILABLE_FOR_USER_IDS": {
			Key:            "ADS_AVAILABLE_FOR_USER_IDS",
			Value:          "0,1",
			Type:           application.ConfigTypeString,
			Description:    "List of user ids for which ads is active",
			Category:       application.ConfigCategoryAd,
			ReleaseVersion: "08.08.22",
		},
	}
}
