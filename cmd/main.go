package main

import (
	"context"
	"crypto/tls"
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/ops"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/shutdown"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/go-common/wrappers/auth_go"
	"github.com/digitalmonsters/go-common/wrappers/follow"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/music/cmd/creator"
	"github.com/digitalmonsters/music/cmd/music"
	"github.com/digitalmonsters/music/configs"
	"github.com/digitalmonsters/music/pkg/creators"
	"github.com/digitalmonsters/music/pkg/creators/notifier"
	feedPkg "github.com/digitalmonsters/music/pkg/feed"
	"github.com/digitalmonsters/music/pkg/feed/deduplicator"
	"github.com/digitalmonsters/music/pkg/feed/feed_converter"
	"github.com/digitalmonsters/music/pkg/global"
	"github.com/digitalmonsters/music/pkg/music_source"
	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"
	apmgoredis "go.elastic.co/apm/module/apmgoredisv8"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// trigger build
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	boilerplate.SetupZeroLog()

	var rootApplication application.RootApplication
	ctx, healthCancel := context.WithCancel(context.Background())

	cfg := configs.GetConfig()
	authGoWrapper := auth_go.NewAuthGoWrapper(cfg.Wrappers.AuthGo)
	userGoWrapper := user_go.NewUserGoWrapper(cfg.Wrappers.UserGo)
	followWrapper := follow.NewFollowWrapper(cfg.Wrappers.Follows)

	httpRouter := router.NewRouter("/rpc", authGoWrapper).
		StartAsync(cfg.HttpPort)

	redisOptions := &redis.Options{
		Addr: cfg.Redis.Host,
		DB:   cfg.Redis.Db,
	}

	if cfg.Redis.Tls {
		redisOptions.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	redisClient := redis.NewClient(redisOptions)
	redisClient.AddHook(apmgoredis.NewHook())

	cfgService := application.NewConfigurator[configs.AppConfig]().
		WithRetriever(application.NewHttpRetriever(application.HttpRetrieverDefaultUrl)).
		WithMigrator(application.NewHttpMigrator(application.HttpMigratorDefaultUrl), configs.GetConfigsMigration()).
		MustInit()

	apiDef := map[string]swagger.ApiDescription{}

	privateRouter := ops.NewPrivateHttpServer().StartAsync(
		cfg.PrivateHttpPort,
	)

	jobber, err := configs.GetJobber(cfg.Jobber)
	if err != nil {
		log.Err(err).Msgf("[Jobber] Could not create jobber")
	}

	musicStorageService := music_source.NewMusicStorageService(&cfg)

	creatorsNotifier := notifier.NewService(
		time.Duration(cfg.NotifierCreatorsConfig.PollTimeMs)*time.Millisecond,
		eventsourcing.NewKafkaEventPublisher(cfg.KafkaWriter, cfg.NotifierCreatorsConfig.KafkaTopic),
		ctx,
	)

	notifiers := []global.INotifier{
		creatorsNotifier,
	}

	creatorsService := creators.NewService(notifiers)

	feedConverter := feed_converter.NewFeedConverter(userGoWrapper, followWrapper, ctx)
	deDuplicator := deduplicator.NewDeDuplicator(redisClient)
	feedService := feedPkg.NewFeed(deDuplicator, feedConverter, jobber, cfgService)

	rootApplication.
		AddApplication(creator.Application(httpRouter, apiDef, creatorsService, userGoWrapper, cfg.Creators, &cfg, feedService)).
		AddApplication(music.Application(httpRouter, apiDef, musicStorageService, &cfg)).
		MustInit()

	if boilerplate.GetCurrentEnvironment() != boilerplate.Prod {
		httpRouter.RegisterDocs(apiDef, nil)
		httpRouter.RegisterProfiler()
	}

	privateRouter.Ready()

	sg := <-sig
	log.Logger.Info().Msgf("GOT SIGNAL %v", sg.String())

	sleepDuration := shutdown.GetGracefulSleepDuration()
	shutdown.RunGracefulShutdown(sleepDuration, []func() error{
		func() error {
			privateRouter.UnHealthy()
			return nil
		},
		func() error {
			healthCancel()
			return nil
		},
		func() error {
			return nil
		},
	})
}
