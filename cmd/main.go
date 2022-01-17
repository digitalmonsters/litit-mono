package main

import (
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/ops"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/shutdown"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/go-common/wrappers/auth"
	"github.com/digitalmonsters/music/cmd/api"
	"github.com/digitalmonsters/music/configs"
	"github.com/digitalmonsters/music/pkg/soundstripe"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// trigger build
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	boilerplate.SetupZeroLog()

	cfg := configs.GetConfig()
	authWrapper := auth.NewAuthWrapper(cfg.Wrappers.Auth)
	httpRouter := router.NewRouter("/rpc", authWrapper).
		StartAsync(cfg.HttpPort)

	apiDef := map[string]swagger.ApiDescription{}

	privateRouter := ops.NewPrivateHttpServer().StartAsync(
		cfg.PrivateHttpPort,
	)

	soundStripeService := soundstripe.NewService(*cfg.SoundStripe)

	if err := api.InitAdminApi(httpRouter, apiDef, soundStripeService); err != nil {
		log.Panic().Err(err).Msg("[Admin API] Cannot initialize api")
		panic(err)
	}

	if err := api.InitPublicApi(httpRouter, apiDef); err != nil {
		log.Panic().Err(err).Msg("[Public API] Cannot initialize api")
		panic(err)
	}
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
			return nil
		},
		func() error {
			return nil
		},
	})
}
