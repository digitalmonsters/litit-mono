package feed

import (
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/music/pkg/feed"
	"github.com/rs/zerolog"
)

type feedApp struct {
	httpRouter       *router.HttpRouter
	apiDef           map[string]swagger.ApiDescription
	musicFeedService *feed.Feed
}

func SubApp(
	httpRouter *router.HttpRouter,
	apiDef map[string]swagger.ApiDescription,
	musicFeedService *feed.Feed,
) application.SubApplication {
	return &feedApp{
		httpRouter:       httpRouter,
		apiDef:           apiDef,
		musicFeedService: musicFeedService,
	}
}

func (f feedApp) Init(subAppLogger zerolog.Logger) error {
	return f.initPublicApi()
}

func (f feedApp) Name() string {
	return "music_feed"
}

func (f feedApp) Close() error {
	return nil
}
