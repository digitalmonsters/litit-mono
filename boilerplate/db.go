package boilerplate

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	postgres "go.elastic.co/apm/module/apmgormv2/driver/postgres"
	"gorm.io/gorm"
	"time"
)

func GetDbConnectionString(config DbConfig) (string, error) {
	return fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v", config.Host, config.User, config.Password, config.Db, config.Port), nil
}

func GetGormConnection(config DbConfig) (*gorm.DB, error) {
	connStr, err := GetDbConnectionString(config)

	if err != nil {
		return nil, err
	}

	mainDb, err := gorm.Open(postgres.Open(connStr), &gorm.Config{
		QueryFields: false,
		Logger: NewDbLogger(CustomLoggerConfig{
			SlowThreshold:         1 * time.Second,
			SkipErrRecordNotFound: true,
			HasStackTrace:         true,
		}),
	})

	if err != nil {
		return nil, errors.WithStack(err)
	}

	rawDb, err := mainDb.DB()

	if err != nil {
		return nil, errors.WithStack(err)
	}

	if config.MaxConnectionLifetimeSec == 0 {
		config.MaxConnectionLifetimeSec = 60
	}

	if config.MaxConnectionIdleSec == 0 {
		config.MaxConnectionLifetimeSec = 50
	}

	if config.MaxIdleConnections == 0 {
		config.MaxIdleConnections = 2
	}

	if config.MaxOpenConnections == 0 {
		config.MaxOpenConnections = 256
	}

	log.Info().Msgf("========= DB [%v] [%v] ============", config.Host, config.Db)

	if config.MaxIdleConnections > config.MaxOpenConnections {
		log.Warn().Msgf("MaxIdleConnections connection should be less or equal to MaxOpenConnections")
		config.MaxIdleConnections = config.MaxOpenConnections
	}

	if config.MaxConnectionIdleSec > config.MaxConnectionLifetimeSec {
		log.Warn().Msgf("MaxConnectionIdleSec should be less or equal to MaxConnectionLifetimeSec")
		config.MaxConnectionIdleSec = config.MaxConnectionLifetimeSec
	}

	log.Info().Msgf("MaxIdleConnections %v", config.MaxOpenConnections)
	log.Info().Msgf("MaxConnectionLifetimeSec %v", config.MaxConnectionLifetimeSec)
	log.Info().Msgf("MaxOpenConnections %v", config.MaxOpenConnections)
	log.Info().Msgf("MaxIdleTimeSec %v", config.MaxConnectionIdleSec)
	log.Info().Msg("=========== DB END ===========")

	rawDb.SetMaxIdleConns(config.MaxIdleConnections)
	rawDb.SetConnMaxLifetime(time.Duration(config.MaxConnectionLifetimeSec) * time.Second)
	rawDb.SetMaxOpenConns(config.MaxOpenConnections)
	rawDb.SetConnMaxIdleTime(time.Duration(config.MaxConnectionIdleSec) * time.Second)

	return mainDb, nil
}
