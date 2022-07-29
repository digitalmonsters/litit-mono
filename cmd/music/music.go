package music

import (
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/music/cmd/music/internal/api"
	"github.com/digitalmonsters/music/configs"
	"github.com/digitalmonsters/music/pkg/music_source"
)

func Application(
	httpRouter *router.HttpRouter,
	apiDef map[string]swagger.ApiDescription,
	musicStorageService *music_source.MusicStorageService,
	cfg *configs.Settings,
) *application.BaseApplication {
	return application.NewBaseApplication("music").
		AddSubApplication(api.SubApp(httpRouter, apiDef, musicStorageService, cfg))
}
