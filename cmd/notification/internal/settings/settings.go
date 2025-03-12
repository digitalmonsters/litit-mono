package settings

import (
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/notification-handler/pkg/settings"
	"github.com/rs/zerolog"
)

type settingsApp struct {
	httpRouter      *router.HttpRouter
	settingsService settings.IService
}

func SubApp(
	httpRouter *router.HttpRouter,
	settingsService settings.IService,
) application.SubApplication {
	return &settingsApp{
		httpRouter:      httpRouter,
		settingsService: settingsService,
	}
}

func (a settingsApp) Init(subAppLogger zerolog.Logger) error {
	if err := a.initPublicApi(a.httpRouter); err != nil {
		return err
	}

	if err := a.initAdminApi(a.httpRouter.GetRpcAdminEndpoint()); err != nil {
		return err
	}

	return nil
}

func (a settingsApp) Name() string {
	return "settings"
}

func (a settingsApp) Close() error {
	return nil
}
