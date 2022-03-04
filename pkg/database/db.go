package database

import (
	"fmt"

	"github.com/digitalmonsters/comments/configs"
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/rs/zerolog/log"
	postgres "go.elastic.co/apm/module/apmgormv2/driver/postgres"
	"gorm.io/gorm"
)

var gormDb *gorm.DB

func init() {
	config := configs.GetConfig()

	db, err := gorm.Open(postgres.Open(fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v",
		config.Db.Host, config.Db.User, config.Db.Password, config.Db.Db, config.Db.Port)), &gorm.Config{
		QueryFields: true,
	})

	if err != nil {
		panic(err)
	}

	gormDb = db

	m := gormigrate.New(db, gormigrate.DefaultOptions, getMigrations())

	log.Info().Msg("start migrations")

	if err = m.Migrate(); err != nil {
		panic(err)
	}
}

func GetDb() *gorm.DB {
	return gormDb
}
