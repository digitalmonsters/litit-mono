package notification

import (
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/notification-handler/cmd/notification/internal/settings"
	settingsPkg "github.com/digitalmonsters/notification-handler/pkg/settings"
)

func Application(
	httpRouter *router.HttpRouter,
	apiDef map[string]swagger.ApiDescription,
	settingsService settingsPkg.IService,
) *application.BaseApplication {
	return application.NewBaseApplication("notification").
		AddSubApplication(settings.SubApp(httpRouter, apiDef, settingsService))
}
