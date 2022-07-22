package moods

import (
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/rs/zerolog"
)

type moodsApp struct {
	httpRouter *router.HttpRouter
	apiDef     map[string]swagger.ApiDescription
}

func SubApp(
	httpRouter *router.HttpRouter,
	apiDef map[string]swagger.ApiDescription,
) application.SubApplication {
	return &moodsApp{
		httpRouter: httpRouter,
		apiDef:     apiDef,
	}
}

func (c moodsApp) Init(subAppLogger zerolog.Logger) error {
	if err := c.initPublicApi(); err != nil {
		return err
	}

	return c.initAdminApi()
}

func (c moodsApp) Name() string {
	return "moods"
}

func (c moodsApp) Close() error {
	return nil
}
