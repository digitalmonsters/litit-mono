package api

import (
	"github.com/digitalmonsters/configurator/pkg/configs"
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/rs/zerolog"
)

type apiApp struct {
	apiDef     map[string]swagger.ApiDescription
	httpRouter *router.HttpRouter
	service    configs.IConfigService
	publisher  eventsourcing.Publisher[eventsourcing.ConfigEvent]
}

func SubApp(
	httpRouter *router.HttpRouter,
	apiDef map[string]swagger.ApiDescription,
	service configs.IConfigService,
	publisher eventsourcing.Publisher[eventsourcing.ConfigEvent],
) application.SubApplication {
	return &apiApp{
		httpRouter: httpRouter,
		apiDef:     apiDef,
		service:    service,
		publisher:  publisher,
	}
}

func (a apiApp) Init(subAppLogger zerolog.Logger) error {
	if err := a.initAdminApi(a.httpRouter.GetRpcAdminEndpoint()); err != nil {
		return err
	}
	a.httpRouter.Router().Handle(a.jsonApi())
	return nil
}

func (a apiApp) Name() string {
	return "api"
}

func (a apiApp) Close() error {
	return nil
}
