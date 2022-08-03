package main

import (
	"context"
	"github.com/RichardKnop/machinery/v1"
	adCampaignApp "github.com/digitalmonsters/ads-manager/cmd/ad_campaign"
	"github.com/digitalmonsters/ads-manager/cmd/api"
	"github.com/digitalmonsters/ads-manager/cmd/common"
	"github.com/digitalmonsters/ads-manager/cmd/consumers/view_content"
	"github.com/digitalmonsters/ads-manager/configs"
	"github.com/digitalmonsters/ads-manager/pkg/ad_campaign"
	"github.com/digitalmonsters/ads-manager/pkg/ad_campaign/ad_moderation"
	commonPkg "github.com/digitalmonsters/ads-manager/pkg/common"
	converter2 "github.com/digitalmonsters/ads-manager/pkg/converter"
	"github.com/digitalmonsters/ads-manager/pkg/database"
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/ops"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/shutdown"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/go-common/wrappers/auth_go"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/digitalmonsters/go-common/wrappers/go_tokenomics"
	"github.com/digitalmonsters/go-common/wrappers/notification_handler"
	"github.com/digitalmonsters/go-common/wrappers/user_category"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	boilerplate.SetupZeroLog()

	healthContext := context.Background()
	var rootApplication application.RootApplication
	cfg := configs.GetConfig()
	authGoWrapper := auth_go.NewAuthGoWrapper(cfg.Wrappers.AuthGo)
	httpRouter := router.NewRouter("/rpc", authGoWrapper).
		StartAsync(cfg.HttpPort)

	apiDef := map[string]swagger.ApiDescription{}

	privateRouter := ops.NewPrivateHttpServer().StartAsync(
		cfg.PrivateHttpPort,
	)

	userGoWrapper := user_go.NewUserGoWrapper(cfg.Wrappers.UserGo)
	contentWrapper := content.NewContentWrapper(cfg.Wrappers.Content)
	userCategoryWrapper := user_category.NewUserCategoryWrapper(cfg.Wrappers.UserCategories)
	notificationHandler := notification_handler.NewNotificationHandlerWrapper(cfg.Wrappers.NotificationHandler)
	goTokenomicsWrapper := go_tokenomics.NewGoTokenomicsWrapper(cfg.Wrappers.GoTokenomics)

	if err := api.InitAdminApi(httpRouter.GetRpcAdminEndpoint(), apiDef); err != nil {
		log.Panic().Err(err).Msg("[Admin API] Cannot initialize api")
		panic(err)
	}

	if err := api.InitPublicApi(httpRouter, apiDef, userGoWrapper); err != nil {
		log.Panic().Err(err).Msg("[Public API] Cannot initialize api")
		panic(err)
	}

	log.Info().Msg("getting jobber")

	jobber, err := configs.GetJobber(&cfg.Jobber)

	if err != nil {
		log.Err(err).Msgf("[Jobber] Could not create jobber")
	}

	_ = jobber.RegisterTask("", func() error {
		return nil
	})

	var machineryWorker *machinery.Worker

	go func() {
		defer func() {
			_ = recover() // https://github.com/RichardKnop/machinery/issues/437
		}()

		machineryWorker = jobber.NewCustomQueueWorker(boilerplate.GetGenerator().Generate().String(),
			cfg.Jobber.Concurrency, cfg.Jobber.DefaultQueue)

		if err = machineryWorker.Launch(); err != nil {
			if err != machinery.ErrWorkerQuitGracefully {
				log.Logger.Err(err).Send()
			}
		}
	}()

	adCampaignService := ad_campaign.NewService(contentWrapper, userCategoryWrapper, userGoWrapper, jobber)
	converter := converter2.NewConverter(userGoWrapper)
	adModerationService := ad_moderation.NewService(notificationHandler, converter)
	commonService := commonPkg.NewService()

	rootApplication.
		AddApplication(adCampaignApp.Application(httpRouter, apiDef, adCampaignService, adModerationService)).
		AddApplication(common.Application(httpRouter, apiDef, commonService)).
		MustInit()

	viewContentListener := view_content.InitListener(healthContext, database.GetDb(database.DbTypeMaster),
		cfg.ViewContentListener, goTokenomicsWrapper).
		ListenAsync()

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
			return viewContentListener.Close()
		},
	})
}
