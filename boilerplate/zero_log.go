package boilerplate

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"math/rand"
	"os"
	"time"
)

func SetupZeroLog() {
	rand.Seed(time.Now().Unix())
	log.Logger = zerolog.New(os.Stderr).With().Caller().Time("time", time.Now().UTC()).Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr, NoColor: true})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}
