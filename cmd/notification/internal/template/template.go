package template

import (
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/notification-handler/pkg/template"
	"github.com/rs/zerolog"
)

type templateApp struct {
	httpRouter      *router.HttpRouter
	apiDef          map[string]swagger.ApiDescription
	templateService template.IService
}

func SubApp(
	httpRouter *router.HttpRouter,
	apiDef map[string]swagger.ApiDescription,
	templateService template.IService,
) application.SubApplication {
	return &templateApp{
		httpRouter:      httpRouter,
		apiDef:          apiDef,
		templateService: templateService,
	}
}

func (a templateApp) Init(subAppLogger zerolog.Logger) error {
	if err := a.initAdminApi(a.httpRouter.GetRpcAdminEndpoint()); err != nil {
		return err
	}

	return nil
}

func (a templateApp) Name() string {
	return "template"
}

func (a templateApp) Close() error {
	return nil
}
