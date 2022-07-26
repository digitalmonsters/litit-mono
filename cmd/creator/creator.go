package creator

import (
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/music/cmd/creator/internal/api"
	"github.com/digitalmonsters/music/cmd/creator/internal/categories"
	"github.com/digitalmonsters/music/cmd/creator/internal/feed"
	"github.com/digitalmonsters/music/cmd/creator/internal/moods"
	"github.com/digitalmonsters/music/cmd/creator/internal/reject_reasons"
	"github.com/digitalmonsters/music/configs"
	"github.com/digitalmonsters/music/pkg/creators"
	feedPkg "github.com/digitalmonsters/music/pkg/feed"
)

func Application(
	httpRouter *router.HttpRouter,
	apiDef map[string]swagger.ApiDescription,
	creatorsService *creators.Service,
	userGoWrapper user_go.IUserGoWrapper,
	creatorsCfg configs.CreatorsConfig,
	cfg *configs.Settings,
	musicFeedService *feedPkg.Feed,
) *application.BaseApplication {
	return application.NewBaseApplication("creator").
		AddSubApplication(api.SubApp(httpRouter, apiDef, creatorsService, userGoWrapper, creatorsCfg, cfg)).
		AddSubApplication(reject_reasons.SubApp(httpRouter, apiDef)).
		AddSubApplication(moods.SubApp(httpRouter, apiDef)).
		AddSubApplication(categories.SubApp(httpRouter, apiDef)).
		AddSubApplication(feed.SubApp(httpRouter, apiDef, musicFeedService))
}
