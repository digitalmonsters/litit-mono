package common

import (
	"github.com/digitalmonsters/ads-manager/cmd/common/internal/api"
	"github.com/digitalmonsters/ads-manager/pkg/common"
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/router"
)

func Application(
	httpRouter *router.HttpRouter,
	commonService common.IService,
) *application.BaseApplication {
	return application.NewBaseApplication("common").
		AddSubApplication(api.SubApp(httpRouter, commonService))
}
