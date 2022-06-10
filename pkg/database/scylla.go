package database

import (
	"context"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/digitalmonsters/go-common/scylla_migrator"
	"github.com/digitalmonsters/notification-handler/configs"
	"github.com/digitalmonsters/notification-handler/pkg/database/scylla_migrations"
	"github.com/gocql/gocql"
	"github.com/rs/zerolog/log"
	"github.com/scylladb/gocqlx/v2"
	"time"
)

var session *gocql.Session

func init() {
	if boilerplate.GetCurrentEnvironment() == boilerplate.Ci || boilerplate.GetCurrentEnvironment() == boilerplate.Local {
		initCi()
	} else {
		initNonCi()
	}
}

func initNonCi() {
	config := configs.GetConfig()
	var err error

	cluster := boilerplate.GetScyllaCluster(config.Scylla)
	session, err = cluster.CreateSession()
	if err != nil {
		log.Panic().Err(err).Send()
	}

	if migrationSession, err := gocqlx.WrapSession(cluster.CreateSession()); err != nil {
		log.Panic().Err(err).Msg("cannot create scylla session")
	} else {
		if err := scylla_migrator.FromFS(context.Background(), migrationSession, scylla_migrations.Files); err != nil {
			log.Err(err).Msg("failed to run migrations")
			panic(err)
		}

		go func() {
			time.Sleep(10 * time.Second)
			migrationSession.Close()
		}()
	}
}

func initCi() {
	config := configs.GetConfig()

	var err error

	_, session, err = boilerplate_testing.GetScyllaTestCluster(&config.Scylla)
	if err != nil {
		log.Panic().Err(err).Msg("cannot initialize scylla session")
		return
	}

	configs.UpdateScyllaKeyspaceForCiConfig(config.Scylla.Keyspace)

	if err := boilerplate_testing.RunScyllaMigrations(session, scylla_migrations.Files); err != nil {
		log.Panic().Err(err).Msg("cannot run migrations")
	}
}

func GetScyllaSession() *gocql.Session {
	return session
}
