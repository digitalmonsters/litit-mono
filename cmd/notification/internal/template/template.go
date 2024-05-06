package template

import (
	"context"
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/go-common/wrappers/content_uploader"
	"github.com/digitalmonsters/notification-handler/pkg/template"

	"github.com/rs/zerolog"
)

type templateApp struct {
	httpRouter      *router.HttpRouter
	apiDef          map[string]swagger.ApiDescription
	templateService template.IService
	uploaderWrapper content_uploader.IContentUploaderWrapper
	appCtx          context.Context
}

func SubApp(
	httpRouter *router.HttpRouter,
	apiDef map[string]swagger.ApiDescription,
	templateService template.IService,
	uploaderWrapper content_uploader.IContentUploaderWrapper,
	appCtx context.Context,
) application.SubApplication {
	return &templateApp{
		httpRouter:      httpRouter,
		apiDef:          apiDef,
		templateService: templateService,
		uploaderWrapper: uploaderWrapper,
		appCtx:          appCtx,
	}
}

func (a templateApp) Init(subAppLogger zerolog.Logger) error {
	if err := a.initAdminApi(a.httpRouter.GetRpcAdminEndpoint()); err != nil {
		return err
	}

	if err := a.initUploaderApi(); err != nil {
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
