package main

import (
	"context"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/ops"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/shutdown"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/go-common/wrappers/auth_go"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/music/cmd/api/creator"
	"github.com/digitalmonsters/music/cmd/api/music"
	"github.com/digitalmonsters/music/configs"
	"github.com/digitalmonsters/music/pkg/creators"
	"github.com/digitalmonsters/music/pkg/creators/notifier"
	"github.com/digitalmonsters/music/pkg/global"
	"github.com/digitalmonsters/music/pkg/music_source"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// trigger build1
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	boilerplate.SetupZeroLog()

	ctx, healthCancel := context.WithCancel(context.Background())

	cfg := configs.GetConfig()
	authGoWrapper := auth_go.NewAuthGoWrapper(cfg.Wrappers.AuthGo)
	userGoWrapper := user_go.NewUserGoWrapper(cfg.Wrappers.UserGo)

	httpRouter := router.NewRouter("/rpc", authGoWrapper).
		StartAsync(cfg.HttpPort)

	apiDef := map[string]swagger.ApiDescription{}

	privateRouter := ops.NewPrivateHttpServer().StartAsync(
		cfg.PrivateHttpPort,
	)

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

	if err := music.InitAdminApi(httpRouter.GetRpcAdminLegacyEndpoint(), apiDef, musicStorageService); err != nil {
		log.Panic().Err(err).Msg("[Music Admin API] Cannot initialize api")
		panic(err)
	}

	if err := music.InitPublicApi(httpRouter, apiDef, musicStorageService); err != nil {
		log.Panic().Err(err).Msg("[Music Public API] Cannot initialize api")
		panic(err)
	}

	if err := creator.InitPublicApi(httpRouter, apiDef, creatorsService, userGoWrapper); err != nil {
		log.Panic().Err(err).Msg("[Creators Public API] Cannot initialize api")
		panic(err)
	}

	if err := creator.InitAdminApi(httpRouter.GetRpcAdminEndpoint(), apiDef, cfg, userGoWrapper, creatorsService); err != nil {
		log.Panic().Err(err).Msg("[Creators Admin API] Cannot initialize api")
		panic(err)
	}

	music.InitUploadApi(httpRouter, &cfg)
	creator.InitUploadApi(httpRouter, &cfg)

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
