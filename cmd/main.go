package main

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/docs"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/shutdown"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/go-common/wrappers/auth"
	"github.com/digitalmonsters/music/cmd/api"
	"github.com/digitalmonsters/music/configs"
	"github.com/digitalmonsters/music/pkg/soundstripe"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	boilerplate.SetupZeroLog()

	cfg := configs.GetConfig()
	healthContext, healthCancel := context.WithCancel(context.Background())
	authWrapper := auth.NewAuthWrapper(cfg.Wrappers.Auth)

	httpRouter := router.NewRouter("/rpc", authWrapper, healthContext)

	apiDef := map[string]swagger.ApiDescription{}

	srv := &fasthttp.Server{
		Handler: fasthttp.CompressHandlerBrotliLevel(httpRouter.Handler(),
			fasthttp.CompressDefaultCompression, fasthttp.CompressDefaultCompression)}

	go func() {
		host := fmt.Sprintf("0.0.0.0:%v", cfg.HttpPort)

		log.Logger.Info().Msgf("[HTTP] Listening on %v", host)

		if err := srv.ListenAndServe(host); err != nil {
			log.Logger.Panic().Err(err).Send()
		}
	}()

	soundStripeService := soundstripe.NewService(*cfg.SoundStripe)

	if err := api.InitAdminApi(httpRouter, apiDef, soundStripeService); err != nil {
		log.Panic().Err(err).Msg("[Admin API] Cannot initialize api")
		panic(err)
	}

	if err := api.InitPublicApi(httpRouter, apiDef); err != nil {
		log.Panic().Err(err).Msg("[Public API] Cannot initialize api")
		panic(err)
	}

	httpRouter.Ready()

	if boilerplate.GetCurrentEnvironment() != boilerplate.Prod {
		var apiCmd []swagger.IApiCommand

		for _, c := range httpRouter.GetRestRegisteredCommands() {
			apiCmd = append(apiCmd, c)
		}

		for _, c := range httpRouter.GetRpcRegisteredCommands() {
			apiCmd = append(apiCmd, c)
		}

		docs.RegisterHttpDoc(httpRouter, "/swagger", apiCmd,
			apiDef, nil)
	}

	sg := <-sig
	log.Logger.Info().Msgf("GOT SIGNAL %v", sg.String())

	sleepDuration := shutdown.GetGracefulSleepDuration()
	shutdown.RunGracefulShutdown(sleepDuration, []func() error{
		func() error {
			httpRouter.NotReady()
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
