package notification

import (
	"context"
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/go-common/wrappers/content_uploader"
	"github.com/digitalmonsters/notification-handler/cmd/notification/internal/settings"
	"github.com/digitalmonsters/notification-handler/cmd/notification/internal/template"
	settingsPkg "github.com/digitalmonsters/notification-handler/pkg/settings"
	templatePkg "github.com/digitalmonsters/notification-handler/pkg/template"
)

func Application(
	httpRouter *router.HttpRouter,
	apiDef map[string]swagger.ApiDescription,
	settingsService settingsPkg.IService,
	templateService templatePkg.IService,
	uploaderWrapper content_uploader.IContentUploaderWrapper,
	appCtx context.Context,
) *application.BaseApplication {
	return application.NewBaseApplication("notification").
		AddSubApplication(settings.SubApp(httpRouter, apiDef, settingsService)).
		AddSubApplication(template.SubApp(httpRouter, apiDef, templateService, uploaderWrapper, appCtx))
}
