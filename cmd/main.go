package main

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/comments/cmd/api/comments"
	"github.com/digitalmonsters/comments/cmd/api/report"
	"github.com/digitalmonsters/comments/cmd/api/vote"
	"github.com/digitalmonsters/comments/configs"
	"github.com/digitalmonsters/comments/pkg/database"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/docs"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/shutdown"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/go-common/wrappers/auth"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/digitalmonsters/go-common/wrappers/user"
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
	db := database.GetDb()
	apiDef := map[string]swagger.ApiDescription{}

	userWrapper := user.NewUserWrapper("")          // todo
	contentWrapper := content.NewContentWrapper("") // todo

	httpRouter := router.NewRouter("/rpc", auth.NewAuthWrapper(cfg.AuthApiUrl))

	if err := comments.Init(httpRouter, db, userWrapper, contentWrapper, apiDef); err != nil {
		panic(err)
	}

	if err := report.Init(httpRouter, db, apiDef); err != nil {
		panic(err)
	}

	if err := vote.Init(httpRouter, db, apiDef); err != nil {
		panic(err)
	}

	if boilerplate.GetCurrentEnvironment() != boilerplate.Prod {
		var apiCmd []swagger.IApiCommand

		for _, c := range httpRouter.GetRestRegisteredCommands() {
			apiCmd = append(apiCmd, c)
		}

		docs.RegisterHttpDoc(httpRouter, "/swagger", apiCmd,
			apiDef, nil)
	}

	healthContext, healthCancel := context.WithCancel(context.Background())
	shutdown.RegisterHttpHealthCheck(healthContext, httpRouter)

	srv := &fasthttp.Server{
		Handler: fasthttp.CompressHandlerBrotliLevel(httpRouter.Handler(),
			fasthttp.CompressDefaultCompression, fasthttp.CompressDefaultCompression),
	}

	go func() {
		host := fmt.Sprintf("0.0.0.0:%v", cfg.HttpPort)

		log.Logger.Info().Msgf("[HTTP] Listening on %v", host)

		if err := srv.ListenAndServe(host); err != nil {
			log.Logger.Panic().Err(err).Send()
		}
	}()

	sg := <-sig
	log.Logger.Info().Msgf("GOT SIGNAL %v", sg.String())

	sleepDuration := shutdown.GetGracefulSleepDuration()
	shutdown.RunGracefulShutdown(sleepDuration, []func() error{
		func() error {
			healthCancel()
			return nil
		},
		func() error {
			return nil
		},
	})
}
