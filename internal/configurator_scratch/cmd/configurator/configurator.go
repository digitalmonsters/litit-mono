package configurator

import (
	"github.com/digitalmonsters/configurator/cmd/configurator/api"
	"github.com/digitalmonsters/configurator/pkg/configs"
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
)

func Application(
	httpRouter *router.HttpRouter,
	apiDef map[string]swagger.ApiDescription,
	service configs.IConfigService,
	publisher eventsourcing.Publisher[eventsourcing.ConfigEvent],
) *application.BaseApplication {
	return application.NewBaseApplication("configurator").
		AddSubApplication(api.SubApp(httpRouter, apiDef, service, publisher))
}
