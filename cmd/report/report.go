package report

import (
	"github.com/digitalmonsters/comments/cmd/report/internal/admin_api"
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
)

func Application(
	httpRouter *router.HttpRouter,
	userWrapper user_go.IUserGoWrapper,
	contentWrapper content.IContentWrapper,
) *application.BaseApplication {
	return application.NewBaseApplication("report").
		AddSubApplication(admin_api.SubApp(httpRouter, userWrapper, contentWrapper))
}
