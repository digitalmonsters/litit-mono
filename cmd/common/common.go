package common

import (
	"github.com/digitalmonsters/ads-manager/cmd/common/internal/api"
	"github.com/digitalmonsters/ads-manager/pkg/common"
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
)

func Application(
	httpRouter *router.HttpRouter,
	apiDef map[string]swagger.ApiDescription,
	commonService common.IService,
) *application.BaseApplication {
	return application.NewBaseApplication("common").
		AddSubApplication(api.SubApp(httpRouter, apiDef, commonService))
}
