package application

import (
	"context"
	"errors"
	"time"

	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ConfiguratorBuilder[T any] struct {
	initialized bool
	retriever   Retriever
	migrator    Migrator
	logger      zerolog.Logger
	interval    time.Duration
}

func NewConfigurator[T any]() *ConfiguratorBuilder[T] {
	logger := log.Logger.With().Str("app", "configurator").Logger()

	return &ConfiguratorBuilder[T]{
		logger:   logger,
		interval: 1 * time.Minute,
	}
}

func (c ConfiguratorBuilder[T]) WithInterval(duration time.Duration) ConfiguratorBuilder[T] {
	c.interval = duration

	return c
}

func (c ConfiguratorBuilder[T]) WithRetriever(retriever Retriever) ConfiguratorBuilder[T] {
	c.retriever = retriever

	return c
}

func (c ConfiguratorBuilder[T]) WithMigrator(migrator Migrator, configsMap map[string]MigrateConfigModel) ConfiguratorBuilder[T] {
	c.migrator = migrator
	c.migrator.SetMigratorMap(configsMap)
	return c
}

func (c ConfiguratorBuilder[T]) MustInit() *Configurator[T] {

	c.logger.Info().Msg("[SERVICE] : configurator initialized")

	if c.initialized {
		c.logger.Panic().Err(errors.New("configuration client already initialized")).Msg("[SERVICE] : configurator failed")
	}

	result := Configurator[T]{builder: c}

	resp, err := c.migrator.Migrate(context.Background())
	if err != nil {
		c.logger.Panic().Err(errors.New("migrate failed - " + err.Error())).Msg("[SERVICE] : configurator failed")
	} else {
		c.logger.Info().Interface("value", resp).Msg("[SERVICE] : configurator migration successful")
	}

	result.init()

	if err := result.Refresh(context.Background()); err != nil {
		c.logger.Panic().Err(errors.New("result failed - " + err.Error())).Msg("[SERVICE] : configurator failed")
	}

	c.logger.Info().Interface("value", result.Values).Msg("[SERVICE] : configurator successful")

	if c.interval > 0 {
		c.logger.Info().Msgf("starting configuration watcher with interval [%v]", c.interval)

		go func() {
			for {
				time.Sleep(c.interval)

				apmTx := apm_helper.StartNewApmTransaction("Refresh", "configurator", nil,
					nil)

				ctx := boilerplate.CreateCustomContext(context.Background(),
					apmTx, log.Logger)

				if err := result.Refresh(ctx); err != nil {
					c.logger.Panic().Err(errors.New("result failed - " + err.Error())).Msg("[SERVICE] : configurator periodic failed")
					apm_helper.LogError(err, ctx)
					apmTx.End()
					continue
				}

				apmTx.Discard()
			}
		}()
	}

	return &result
}
