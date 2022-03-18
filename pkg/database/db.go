package database

import (
	"github.com/digitalmonsters/comments/configs"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

var gormDb *gorm.DB

func init() {
	config := configs.GetConfig()

	if boilerplate.GetCurrentEnvironment() == boilerplate.Ci {
		if err := boilerplate_testing.EnsurePostgresDbExists(config.Db); err != nil {
			panic(err)
		}
	}

	log.Info().Msg("setup postgres database")

	db, err := boilerplate.GetGormConnection(config.Db)

	if err != nil {
		panic(err)
	}

	gormDb = db

	m := gormigrate.New(gormDb, gormigrate.DefaultOptions, getMigrations())

	log.Info().Msg("[Db] start migrations")

	if err = m.Migrate(); err != nil {
		panic(err)
	}
}

func GetDb() *gorm.DB {
	return gormDb
}
