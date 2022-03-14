package main

import (
	"context"
	"github.com/digitalmonsters/go-common/wrappers/notification_gateway"
	"github.com/digitalmonsters/notification-handler/cmd/consumers/creators"
	"github.com/digitalmonsters/notification-handler/cmd/consumers/sending_queue"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"os"
	"os/signal"
	"syscall"

	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/ops"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/shutdown"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/go-common/wrappers/auth_go"
	"github.com/digitalmonsters/notification-handler/cmd/api/creator"
	"github.com/digitalmonsters/notification-handler/configs"
	"github.com/rs/zerolog/log"
)

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	boilerplate.SetupZeroLog()

	cfg := configs.GetConfig()
	authGoWrapper := auth_go.NewAuthGoWrapper(cfg.Wrappers.AuthGo)
	httpRouter := router.NewRouter("/rpc", authGoWrapper).
		StartAsync(cfg.HttpPort)

	apiDef := map[string]swagger.ApiDescription{}

	privateRouter := ops.NewPrivateHttpServer().StartAsync(
		cfg.PrivateHttpPort,
	)

	ctx := context.Background()

	notificationSender := sender.NewSender(notification_gateway.NewNotificationGatewayWrapper(
		cfg.Wrappers.NotificationGateway))

	sendingQueueListener := sending_queue.InitListener(ctx, cfg.SendingQueueListener, notificationSender).ListenAsync()

	creatorsListener := creators.InitListener(ctx, cfg.CreatorsListener, notificationSender).ListenAsync()

	if err := creator.InitAdminApi(httpRouter.GetRpcAdminLegacyEndpoint(), apiDef, cfg); err != nil {
		log.Panic().Err(err).Msg("cannot initialize api")
		panic(err)
	}

	if boilerplate.GetCurrentEnvironment() != boilerplate.Prod {
		httpRouter.RegisterDocs(apiDef, nil)
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
			return sendingQueueListener.Close()
		},
		func() error {
			return creatorsListener.Close()
		},
	})
}
