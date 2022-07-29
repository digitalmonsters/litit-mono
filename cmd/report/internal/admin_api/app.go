package admin_api

import (
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/rs/zerolog"
)

type adminApiApp struct {
	httpRouter     *router.HttpRouter
	apiDef         map[string]swagger.ApiDescription
	userWrapper    user_go.IUserGoWrapper
	contentWrapper content.IContentWrapper
}

func SubApp(
	httpRouter *router.HttpRouter,
	apiDef map[string]swagger.ApiDescription,
	userWrapper user_go.IUserGoWrapper,
	contentWrapper content.IContentWrapper,
) application.SubApplication {
	return &adminApiApp{
		httpRouter:     httpRouter,
		apiDef:         apiDef,
		userWrapper:    userWrapper,
		contentWrapper: contentWrapper,
	}
}

func (a *adminApiApp) Init(subAppLogger zerolog.Logger) error {
	if err := a.initAdminApi(a.httpRouter.GetRpcAdminEndpoint()); err != nil {
		return err
	}

	return nil
}

func (a *adminApiApp) Name() string {
	return "report_admin_api"
}

func (a *adminApiApp) Close() error {
	return nil
}
