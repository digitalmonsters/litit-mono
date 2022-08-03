package creator

import (
	"context"
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/music/cmd/creator/internal/api"
	"github.com/digitalmonsters/music/cmd/creator/internal/categories"
	"github.com/digitalmonsters/music/cmd/creator/internal/consumers/dislike"
	"github.com/digitalmonsters/music/cmd/creator/internal/consumers/like"
	"github.com/digitalmonsters/music/cmd/creator/internal/consumers/listen"
	"github.com/digitalmonsters/music/cmd/creator/internal/consumers/listened_music"
	"github.com/digitalmonsters/music/cmd/creator/internal/consumers/love"
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
	contentWrapper content.IContentWrapper,
	creatorsCfg configs.CreatorsConfig,
	cfg *configs.Settings,
	musicFeedService *feedPkg.Feed,
	ctx context.Context,
	appConfig *application.Configurator[configs.AppConfig],
) *application.BaseApplication {
	return application.NewBaseApplication("creator").
		AddSubApplication(api.SubApp(httpRouter, apiDef, creatorsService, userGoWrapper, contentWrapper, creatorsCfg, cfg, appConfig)).
		AddSubApplication(reject_reasons.SubApp(httpRouter, apiDef)).
		AddSubApplication(moods.SubApp(httpRouter, apiDef)).
		AddSubApplication(categories.SubApp(httpRouter, apiDef)).
		AddSubApplication(feed.SubApp(httpRouter, apiDef, musicFeedService)).
		AddSubApplication(like.SubApp(ctx, creatorsCfg.Listeners.LikeCounter)).
		AddSubApplication(dislike.SubApp(ctx, creatorsCfg.Listeners.DislikeCounter)).
		AddSubApplication(love.SubApp(ctx, creatorsCfg.Listeners.LoveCounter)).
		AddSubApplication(listen.SubApp(ctx, creatorsCfg.Listeners.ListenCounter)).
		AddSubApplication(listened_music.SubApp(ctx, creatorsCfg.Listeners.ListenedMusic))
}
