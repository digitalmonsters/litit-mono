package database

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/music/configs"
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/rs/zerolog/log"
	postgres "go.elastic.co/apm/module/apmgormv2/driver/postgres"
	"gorm.io/gorm"
)

type DbType int

const (
	DbTypeMaster   = DbType(1)
	DbTypeReadonly = DbType(2)
)

var masterGormDb *gorm.DB
var readonlyGormDb *gorm.DB

func init() {
	config := configs.GetConfig()

	log.Info().Msg("setup postgres database")

	mainDb, err := gorm.Open(postgres.Open(fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v",
		config.MasterDb.Host, config.MasterDb.User, config.MasterDb.Password, config.MasterDb.Db, config.MasterDb.Port)), &gorm.Config{
		QueryFields: true,
	})

	if err != nil {
		log.Info().Msg(config.MasterDb.Password)
		panic(err)
	}

	masterGormDb = mainDb

	readOnlyConfig := config.ReadonlyDb

	if len(readOnlyConfig.Host) == 0 || len(readOnlyConfig.Db) == 0 {
		readOnlyConfig = config.MasterDb
		log.Warn().Msgf("[DB Connection] no configuration for read-only db. Will use Master DB")
	}

	readDb, err := gorm.Open(postgres.Open(fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v",
		readOnlyConfig.Host, readOnlyConfig.User, readOnlyConfig.Password, readOnlyConfig.Db, readOnlyConfig.Port)), &gorm.Config{
		QueryFields: true,
	})

	if err != nil {
		log.Err(err).Msg("[Db Connection] can not setup connection to read-only db, will use master")

		readonlyGormDb = masterGormDb
	} else {
		readonlyGormDb = readDb
	}

	m := gormigrate.New(mainDb, gormigrate.DefaultOptions, getMigrations())

	log.Info().Msg("[Db] start migrations")

	if err = m.Migrate(); err != nil {
		panic(err)
	}
}

func GetDb(t DbType) *gorm.DB {
	switch t {
	case DbTypeMaster:
		return masterGormDb
	case DbTypeReadonly:
		return readonlyGormDb
	default:
		return masterGormDb
	}
}

func GetDbWithContext(t DbType, ctx context.Context) *gorm.DB {
	return GetDb(t).WithContext(ctx)
}
