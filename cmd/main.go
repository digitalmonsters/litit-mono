package main

import (
	"context"
	"github.com/digitalmonsters/comments/cmd/api"
	"github.com/digitalmonsters/comments/cmd/api/comments"
	"github.com/digitalmonsters/comments/cmd/api/comments/notifiers/comment"
	"github.com/digitalmonsters/comments/cmd/api/comments/notifiers/content_comments_counter"
	"github.com/digitalmonsters/comments/cmd/api/comments/notifiers/user_comments_counter"
	"github.com/digitalmonsters/comments/cmd/api/report"
	"github.com/digitalmonsters/comments/cmd/api/vote"
	vote2 "github.com/digitalmonsters/comments/cmd/api/vote/notifiers/vote"
	"github.com/digitalmonsters/comments/cmd/consumers/user_consumer"
	"github.com/digitalmonsters/comments/configs"
	"github.com/digitalmonsters/comments/pkg/database"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/ops"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/shutdown"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/go-common/wrappers/auth_go"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/digitalmonsters/go-common/wrappers/user_block"
	user "github.com/digitalmonsters/go-common/wrappers/user_go"
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

	ctx, cancel := context.WithCancel(context.Background())
	cfg := configs.GetConfig()
	db := database.GetDb()
	apiDef := map[string]swagger.ApiDescription{}
	fastHttpRouter := router.NewRouter("/rpc", auth_go.NewAuthGoWrapper(cfg.Wrappers.AuthGo)).
		StartAsync(cfg.HttpPort)

	privateRouter := ops.NewPrivateHttpServer().StartAsync(
		cfg.PrivateHttpPort,
	)

	userWrapper := user.NewUserGoWrapper(cfg.Wrappers.UserGo)
	contentWrapper := content.NewContentWrapper(cfg.Wrappers.Content)
	userBlockWrapper := user_block.NewUserBlockWrapper(cfg.Wrappers.UserBlock)

	commentNotifier := comment.NewNotifier(time.Duration(cfg.NotifierCommentConfig.PollTimeMs)*time.Millisecond,
		ctx, eventsourcing.NewKafkaEventPublisher(*cfg.KafkaWriter, cfg.NotifierCommentConfig.KafkaTopic), db, true)
	contentCommentsNotifier := content_comments_counter.NewNotifier(time.Duration(cfg.NotifierContentCommentsCounterConfig.PollTimeMs)*time.Millisecond,
		ctx, eventsourcing.NewKafkaEventPublisher(*cfg.KafkaWriter, cfg.NotifierContentCommentsCounterConfig.KafkaTopic), true)
	userCommentsNotifier := user_comments_counter.NewNotifier(time.Duration(cfg.NotifierUserCommentsCounterConfig.PollTimeMs)*time.Millisecond,
		ctx, eventsourcing.NewKafkaEventPublisher(*cfg.KafkaWriter, cfg.NotifierUserCommentsCounterConfig.KafkaTopic), true)
	voteNotifier := vote2.NewNotifier(time.Duration(cfg.NotifierVoteConfig.PollTimeMs)*time.Millisecond,
		ctx, eventsourcing.NewKafkaEventPublisher(*cfg.KafkaWriter, cfg.NotifierVoteConfig.KafkaTopic), true)

	user_consumer.InitListener(ctx, cfg.UserListener, commentNotifier, contentCommentsNotifier, userCommentsNotifier).
		ListenAsync()

	if err := comments.Init(fastHttpRouter, db, userWrapper, contentWrapper, userBlockWrapper, apiDef, commentNotifier,
		contentCommentsNotifier, userCommentsNotifier); err != nil {
		panic(err)
	}

	if err := report.Init(fastHttpRouter, db, apiDef); err != nil {
		panic(err)
	}

	if err := vote.Init(fastHttpRouter, db, apiDef, commentNotifier, voteNotifier, contentWrapper); err != nil {
		panic(err)
	}

	if err := api.InitInternalApi(fastHttpRouter.GetRpcServiceEndpoint(), apiDef, db); err != nil {
		panic(err)
	}

	if boilerplate.GetCurrentEnvironment() != boilerplate.Prod {
		fastHttpRouter.RegisterDocs(apiDef, nil)
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
			cancel()
			return nil
		},
		func() error {
			return nil
		},
	})
}
