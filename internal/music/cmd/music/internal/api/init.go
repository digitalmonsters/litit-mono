package api

import (
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/music/configs"
	"github.com/digitalmonsters/music/pkg/music_source"
	"github.com/rs/zerolog"
)

type musicApp struct {
	httpRouter          *router.HttpRouter
	apiDef              map[string]swagger.ApiDescription
	musicStorageService *music_source.MusicStorageService
	cfg                 *configs.Settings
	appConfig           *application.Configurator[configs.AppConfig]
}

func SubApp(
	httpRouter *router.HttpRouter,
	apiDef map[string]swagger.ApiDescription,
	musicStorageService *music_source.MusicStorageService,
	cfg *configs.Settings,
	appConfig *application.Configurator[configs.AppConfig],
) application.SubApplication {
	return &musicApp{
		httpRouter:          httpRouter,
		apiDef:              apiDef,
		musicStorageService: musicStorageService,
		cfg:                 cfg,
		appConfig:           appConfig,
	}
}

func (m musicApp) Init(subAppLogger zerolog.Logger) error {
	if err := m.initPublicApi(); err != nil {
		return err
	}

	if err := m.initAdminLegacyApi(); err != nil {
		return err
	}

	m.InitUploadApi()

	return m.initAdminApi()
}

func (m musicApp) Name() string {
	return "music"
}

func (m musicApp) Close() error {
	return nil
}
