package api

import (
	"github.com/digitalmonsters/ads-manager/pkg/common"
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/router"
	"github.com/rs/zerolog"
)

type apiApp struct {
	httpRouter    *router.HttpRouter
	commonService common.IService
}

func SubApp(
	httpRouter *router.HttpRouter,
	commonService common.IService,
) application.SubApplication {
	return &apiApp{
		httpRouter:    httpRouter,
		commonService: commonService,
	}
}

func (a *apiApp) Init(subAppLogger zerolog.Logger) error {
	if err := a.initAdminApi(a.httpRouter.GetRpcAdminEndpoint()); err != nil {
		return err
	}

	if err := a.initPublicApi(a.httpRouter); err != nil {
		return err
	}

	return nil
}

func (a *apiApp) Name() string {
	return "api"
}

func (a *apiApp) Close() error {
	return nil
}
