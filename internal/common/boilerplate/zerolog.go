package boilerplate

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.elastic.co/apm"
	"go.elastic.co/apm/module/apmzerolog"
	"strings"
)

var isConfigured bool

func SetupZeroLog() {
	if isConfigured {
		return
	}

	zerolog.CallerMarshalFunc = func(file string, line int) string {
		sp := strings.Split(file, "/")

		segments := 4

		if len(sp) == 0 { // just in case
			segments = 0
		}

		if segments > 0 && segments > len(sp) {
			segments = len(sp) - 1
		}

		return fmt.Sprintf("%s:%v", strings.Join(sp[segments:], "/"), line)
	}
	log.Logger = log.Logger.With().Caller().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.DefaultContextLogger = &log.Logger

	isConfigured = true
}

func CreateCustomContext(ctx context.Context, apmTx *apm.Transaction, logger zerolog.Logger) context.Context {
	if apmTx != nil {
		ctx = apm.ContextWithTransaction(ctx, apmTx)
		logger = logger.Hook(apmzerolog.TraceContextHook(ctx))
	}

	ctx = logger.WithContext(ctx)

	return ctx
}
