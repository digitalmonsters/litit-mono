package main

import (
	"context"
	"github.com/digitalmonsters/go-common/wrappers/comment"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/digitalmonsters/go-common/wrappers/follow"
	"github.com/digitalmonsters/go-common/wrappers/notification_gateway"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/notification-handler/cmd/api"
	commentConsumer "github.com/digitalmonsters/notification-handler/cmd/consumers/comment"
	contentConsumer "github.com/digitalmonsters/notification-handler/cmd/consumers/content"
	"github.com/digitalmonsters/notification-handler/cmd/consumers/creators"
	"github.com/digitalmonsters/notification-handler/cmd/consumers/email_notification"
	followConsumer "github.com/digitalmonsters/notification-handler/cmd/consumers/follow"
	"github.com/digitalmonsters/notification-handler/cmd/consumers/kyc_status"
	"github.com/digitalmonsters/notification-handler/cmd/consumers/like"
	"github.com/digitalmonsters/notification-handler/cmd/consumers/push_admin_message"
	"github.com/digitalmonsters/notification-handler/cmd/consumers/sending_queue"
	"github.com/digitalmonsters/notification-handler/cmd/consumers/tokenomics_notification"
	"github.com/digitalmonsters/notification-handler/cmd/consumers/user_banned"
	"github.com/digitalmonsters/notification-handler/cmd/consumers/user_delete"
	"github.com/digitalmonsters/notification-handler/cmd/consumers/vote"
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
	//trigger build1
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
	sendingQueueCustomListener := sending_queue.InitListener(ctx, cfg.SendingQueueCustomListener, notificationSender).ListenAsync()

	creatorsListener := creators.InitListener(ctx, cfg.CreatorsListener, notificationSender).ListenAsync()

	userGoWrapper := user_go.NewUserGoWrapper(cfg.Wrappers.UserGo)
	contentWrapper := content.NewContentWrapper(cfg.Wrappers.Content)
	followWrapper := follow.NewFollowWrapper(cfg.Wrappers.Follows)
	commentWrapper := comment.NewCommentWrapper(cfg.Wrappers.Comment)

	commentListener := commentConsumer.InitListener(ctx, cfg.CommentListener, notificationSender, userGoWrapper,
		contentWrapper, commentWrapper).ListenAsync()
	voteListener := vote.InitListener(ctx, cfg.VoteListener, notificationSender, userGoWrapper).ListenAsync()
	likeListener := like.InitListener(ctx, cfg.LikeListener, notificationSender, userGoWrapper, contentWrapper).ListenAsync()
	contentListener := contentConsumer.InitListener(ctx, cfg.ContentListener, notificationSender, followWrapper,
		userGoWrapper, contentWrapper).ListenAsync()
	kycStatusListener := kyc_status.InitListener(ctx, cfg.KysStatusListener, notificationSender).ListenAsync()
	followListener := followConsumer.InitListener(ctx, cfg.FollowListener, notificationSender, userGoWrapper).ListenAsync()
	tokenomicsNotificationListener := tokenomics_notification.InitListener(ctx, cfg.TokenomicsNotificationListener,
		notificationSender, userGoWrapper).ListenAsync()
	emailNotificationListener := email_notification.InitListener(ctx, cfg.EmailNotificationListener,
		notificationSender, userGoWrapper, cfg.EmailLinks).ListenAsync()
	pushAdminMessageListener := push_admin_message.InitListener(ctx, cfg.PushAdminMessageListener,
		notificationSender).ListenAsync()
	userDeleteListener := user_delete.InitListener(ctx, cfg.UserDeleteListener).ListenAsync()
	userBannedListener := user_banned.InitListener(ctx, cfg.UserBannedListener, notificationSender).ListenAsync()

	api.InitInternalApi(httpRouter.GetRpcServiceEndpoint())

	if err := creator.InitAdminApi(httpRouter.GetRpcAdminLegacyEndpoint(), apiDef, cfg); err != nil {
		log.Fatal().Err(err).Msgf("[HTTP] Could not init admin creator api")
	}

	if err := api.InitNotificationApi(httpRouter, apiDef, userGoWrapper, followWrapper); err != nil {
		log.Fatal().Err(err).Msgf("[HTTP] Could not init notification api")
	}

	if err := api.InitAdminNotificationApi(httpRouter, apiDef, userGoWrapper, followWrapper); err != nil {
		log.Fatal().Err(err).Msgf("[HTTP] Could not init admin notification api")
	}

	if err := api.InitTokenApi(httpRouter, apiDef); err != nil {
		log.Fatal().Err(err).Msgf("[HTTP] Could not init token api")
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
			return sendingQueueCustomListener.Close()
		},
		func() error {
			return creatorsListener.Close()
		},
		func() error {
			return commentListener.Close()
		},
		func() error {
			return voteListener.Close()
		},
		func() error {
			return likeListener.Close()
		},
		func() error {
			return contentListener.Close()
		},
		func() error {
			return kycStatusListener.Close()
		},
		func() error {
			return followListener.Close()
		},
		func() error {
			return tokenomicsNotificationListener.Close()
		},
		func() error {
			return emailNotificationListener.Close()
		},
		func() error {
			return pushAdminMessageListener.Close()
		},
		func() error {
			return userDeleteListener.Close()
		},
		func() error {
			return userBannedListener.Close()
		},
	})
}
