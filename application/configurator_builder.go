package application

import (
	"context"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"time"
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
		interval: 10 * time.Second,
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
	if c.initialized {
		panic("configuration client already initialized")
	}

	result := Configurator[T]{builder: c}

	if _, err := c.migrator.Migrate(context.Background()); err != nil {
		panic("cannot migrate config values")
	}

	result.init()

	if err := result.Refresh(context.Background()); err != nil {
		panic(err)
	}

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
