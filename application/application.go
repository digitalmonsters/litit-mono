package application

import (
	"context"
	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"time"
)

type RootApplication struct {
	applications []*BaseApplication
}

func (b *RootApplication) AddApplication(application *BaseApplication) *RootApplication {
	b.applications = append(b.applications, application)

	return b
}

func (b *RootApplication) MustInit() {
	for _, a := range b.applications {
		now := time.Now()
		a.logger.Info().Msgf("[root] application [%v] starting", a.Name())
		a.MustInit()
		a.logger.Info().Msgf("[root] application [%v] started in %v", a.Name(), time.Since(now).String())
	}
}

func NewBaseApplication(appName string) *BaseApplication {
	return &BaseApplication{
		logger: log.Logger.With().Str("app_name", appName).Logger(),
		name:   appName,
	}
}

type BaseApplication struct {
	logger zerolog.Logger
	subApp []SubApplication
	name   string
}

type BaseSubApplication struct {
	logger zerolog.Logger
}

func (b *BaseApplication) MustInit() {
	var final error

	for _, sub := range b.subApp {
		logger := b.logger.With().Str("sub_app", sub.Name()).Logger()
		now := time.Now()
		logger.Info().Msgf("[root] sub application [%v] of [%v] starting", sub.Name(), b.Name())

		if err := sub.Init(logger); err != nil {
			final = multierror.Append(final, err)

			logger.Err(err).Send()
		} else {
			logger.Info().Msgf("[root] sub application [%v] of [%v] started in %v", sub.Name(), b.Name(),
				time.Since(now).String())
		}
	}

	if final != nil {
		panic(final)
	}
}

func NewBaseSubApplication(subAppName string) *BaseSubApplication {
	return &BaseSubApplication{
		logger: log.Logger.With().Str("sub_app_name", subAppName).Logger(),
	}
}

func (b *BaseApplication) AddSubApplication(subApp SubApplication) *BaseApplication {
	b.subApp = append(b.subApp, subApp)

	return b
}

func (b BaseApplication) Name() string {
	return b.name
}

func (b *BaseApplication) Close() error {
	var err error

	for _, sub := range b.subApp {
		func() {
			defer RecoverFunc(context.Background())

			if e := sub.Close(); e != nil {
				err = multierror.Append(err, e)
			}
		}()

	}

	return err
}

type SubApplication interface {
	Init(subAppLogger zerolog.Logger) error
	Name() string
	Close() error
}
