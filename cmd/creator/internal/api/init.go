package api

import (
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/music/configs"
	"github.com/digitalmonsters/music/pkg/creators"
	"github.com/rs/zerolog"
)

type creatorApp struct {
	httpRouter      *router.HttpRouter
	apiDef          map[string]swagger.ApiDescription
	creatorsService *creators.Service
	userGoWrapper   user_go.IUserGoWrapper
	creatorsCfg     configs.CreatorsConfig
	cfg             *configs.Settings
}

func SubApp(
	httpRouter *router.HttpRouter,
	apiDef map[string]swagger.ApiDescription,
	creatorsService *creators.Service,
	userGoWrapper user_go.IUserGoWrapper,
	creatorsCfg configs.CreatorsConfig,
	cfg *configs.Settings,
) application.SubApplication {
	return &creatorApp{
		httpRouter:      httpRouter,
		apiDef:          apiDef,
		creatorsService: creatorsService,
		userGoWrapper:   userGoWrapper,
		creatorsCfg:     creatorsCfg,
		cfg:             cfg,
	}
}

func (c creatorApp) Init(subAppLogger zerolog.Logger) error {
	if err := c.initPublicApi(); err != nil {
		return err
	}

	c.initUploadApi()

	return c.initAdminApi()
}

func (c creatorApp) Name() string {
	return "creators_api"
}

func (c creatorApp) Close() error {
	return nil
}
