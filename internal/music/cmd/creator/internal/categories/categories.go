package categories

import (
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/rs/zerolog"
)

type categoriesApp struct {
	httpRouter *router.HttpRouter
	apiDef     map[string]swagger.ApiDescription
}

func SubApp(
	httpRouter *router.HttpRouter,
	apiDef map[string]swagger.ApiDescription,
) application.SubApplication {
	return &categoriesApp{
		httpRouter: httpRouter,
		apiDef:     apiDef,
	}
}

func (c categoriesApp) Init(subAppLogger zerolog.Logger) error {
	if err := c.initPublicApi(); err != nil {
		return err
	}

	return c.initAdminApi()
}

func (c categoriesApp) Name() string {
	return "categories"
}

func (c categoriesApp) Close() error {
	return nil
}
