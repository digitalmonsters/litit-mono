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
	MUSIC_MAX_HASHTAGS_COUNT                    int
	MUSIC_FEED_LIMIT                            int
	MUSIC_CALCULATION_LOVE_COUNT_WEIGHT         int
	MUSIC_CALCULATION_LIKE_COUNT_WEIGHT         int
	MUSIC_CALCULATION_SHORT_LISTEN_COUNT_WEIGHT int
	MUSIC_CALCULATION_DISLIKE_COUNT_WEIGHT      int
	MUSIC_CALCULATION_TIMING_START_CONF         int
	MUSIC_CALCULATION_TIMING_DELIMITER          int
	MUSIC_FEED_UPDATE_SCORE_FREQUENCY_MINUTES   int
	MUSIC_SHORT_VERSION_MAX_DURATION            int
	MUSIC_FULL_VERSION_MAX_DURATION             int
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
		"MUSIC_FEED_LIMIT": {
			Key:            "MUSIC_FEED_LIMIT",
			Value:          fmt.Sprint(10000),
			Type:           application.ConfigTypeInteger,
			Description:    "Max music feed limit",
			Category:       application.ConfigCategoryApplications,
			ReleaseVersion: "22.07.22",
		},
		"MUSIC_CALCULATION_LOVE_COUNT_WEIGHT": {
			Key:            "MUSIC_CALCULATION_LOVE_COUNT_WEIGHT",
			Value:          "10",
			Type:           application.ConfigTypeInteger,
			Description:    "Music feed. c.loves * [X]",
			Category:       application.ConfigMusic,
			ReleaseVersion: "29.07.2022",
		},
		"MUSIC_CALCULATION_LIKE_COUNT_WEIGHT": {
			Key:            "MUSIC_CALCULATION_LIKE_COUNT_WEIGHT",
			Value:          "6",
			Type:           application.ConfigTypeInteger,
			Description:    "Music feed. c.likes * [X]",
			Category:       application.ConfigMusic,
			ReleaseVersion: "29.07.2022",
		},
		"MUSIC_CALCULATION_DISLIKE_COUNT_WEIGHT": {
			Key:            "MUSIC_CALCULATION_DISLIKE_COUNT_WEIGHT",
			Value:          "1",
			Type:           application.ConfigTypeInteger,
			Description:    "Music feed. (calculations...) - c.dislikes * [X]",
			Category:       application.ConfigMusic,
			ReleaseVersion: "29.07.2022",
		},
		"MUSIC_CALCULATION_SHORT_LISTEN_COUNT_WEIGHT": {
			Key:            "MUSIC_CALCULATION_SHORT_LISTEN_COUNT_WEIGHT",
			Value:          "1",
			Type:           application.ConfigTypeInteger,
			Description:    "Music feed. c.short_listens * [X]",
			Category:       application.ConfigMusic,
			ReleaseVersion: "29.07.2022",
		},
		"MUSIC_CALCULATION_TIMING_START_CONF": {
			Key:            "MUSIC_CALCULATION_TIMING_START_CONF",
			Value:          "5000",
			Type:           application.ConfigTypeInteger,
			Description:    "Music calculation timing start value",
			Category:       application.ConfigMusic,
			ReleaseVersion: "29.07.2022",
		},
		"MUSIC_CALCULATION_TIMING_DELIMITER": {
			Key:            "MUSIC_CALCULATION_TIMING_DELIMITER",
			Value:          "60",
			Type:           application.ConfigTypeInteger,
			Description:    "Music calculation delimiter",
			Category:       application.ConfigMusic,
			ReleaseVersion: "29.07.2022",
		},
		"MUSIC_FEED_UPDATE_SCORE_FREQUENCY_MINUTES": {
			Key:            "MUSIC_FEED_UPDATE_SCORE_FREQUENCY_MINUTES",
			Value:          "60",
			Type:           application.ConfigTypeInteger,
			Description:    "Period per music score update",
			Category:       application.ConfigMusic,
			ReleaseVersion: "29.07.2022",
		},
		"MUSIC_SHORT_VERSION_MAX_DURATION": {
			Key:            "MUSIC_SHORT_VERSION_MAX_DURATION",
			Value:          "15",
			Type:           application.ConfigTypeInteger,
			Description:    "short version of creator song max duration",
			Category:       application.ConfigMusic,
			ReleaseVersion: "05.05.2022",
		},
		"MUSIC_FULL_VERSION_MAX_DURATION": {
			Key:            "MUSIC_FULL_VERSION_MAX_DURATION",
			Value:          "600",
			Type:           application.ConfigTypeInteger,
			Description:    "full version of creator song max duration",
			Category:       application.ConfigMusic,
			ReleaseVersion: "05.05.2022",
		},
	}
}
