package main

import (
	"context"
	"github.com/digitalmonsters/configurator/cmd/configurator"
	"github.com/digitalmonsters/configurator/configs"
	configsPkg "github.com/digitalmonsters/configurator/pkg/configs"
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/ops"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/shutdown"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/go-common/wrappers/auth_go"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	sig := make(chan os.Signal, 2)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	boilerplate.SetupZeroLog()
	config := configs.GetConfig()

	authGoWrapper := auth_go.NewAuthGoWrapper(config.Wrappers.AuthGo)
	fastHttpRouter := router.NewRouter("/rpc", authGoWrapper).
		StartAsync(config.HttpPort)
	privateRouter := ops.NewPrivateHttpServer().StartAsync(
		config.PrivateHttpPort,
	)
	apiDescription := map[string]swagger.ApiDescription{}
	var rootApplication application.RootApplication

	var configPublisher = eventsourcing.NewKafkaBatchPublisher[configsPkg.ConfigEvent]("config_upsert", config.ConfigNotifier, context.Background())

	configService := configsPkg.ConfigService{}
	rootApplication.
		AddApplication(configurator.Application(fastHttpRouter, apiDescription, configService, configPublisher)).
		MustInit()

	log.Info().Msg("bootstrapping configurator")

	if boilerplate.GetCurrentEnvironment() != boilerplate.Prod {
		fastHttpRouter.RegisterDocs(apiDescription, []swagger.ConstantDescription{})
	}

	fastHttpRouter.RegisterProfiler()

	privateRouter.Ready()

	sg := <-sig
	log.Logger.Info().Msgf("GOT SIGNAL %v", sg.String())

	sleepDuration := shutdown.GetGracefulSleepDuration()

	shutdown.RunGracefulShutdown(sleepDuration, []func() error{
		func() error {
			privateRouter.UnHealthy()
			return nil
		},
	})

}
