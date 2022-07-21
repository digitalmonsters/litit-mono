package reject_reasons

import (
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/rs/zerolog"
)

type rejectReasonsApp struct {
	httpRouter *router.HttpRouter
	apiDef     map[string]swagger.ApiDescription
}

func SubApp(
	httpRouter *router.HttpRouter,
	apiDef map[string]swagger.ApiDescription,
) application.SubApplication {
	return &rejectReasonsApp{
		httpRouter: httpRouter,
		apiDef:     apiDef,
	}
}

func (c rejectReasonsApp) Init(subAppLogger zerolog.Logger) error {
	return c.initAdminApi()
}

func (c rejectReasonsApp) Name() string {
	return "reject_reasons"
}

func (c rejectReasonsApp) Close() error {
	return nil
}
