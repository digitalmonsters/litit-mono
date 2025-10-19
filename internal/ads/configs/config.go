package configs

import (
	"crypto/tls"
	"fmt"

	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	"github.com/digitalmonsters/go-common/boilerplate"
)

var CDN_BASE string

const (
	PREFIX_CONTENT = "content"
)

type Settings struct {
	CdnBase             string                                 `json:"CdnBase"`
	HttpPort            int                                    `json:"HttpPort"`
	Wrappers            boilerplate.Wrappers                   `json:"Wrappers"`
	MasterDb            boilerplate.DbConfig                   `json:"MasterDb"`
	ReadonlyDb          boilerplate.DbConfig                   `json:"ReadonlyDb"`
	PrivateHttpPort     int                                    `json:"PrivateHttpPort"`
	Jobber              JobberConfig                           `json:"Jobber"`
	ViewContentListener boilerplate.KafkaListenerConfiguration `json:"ViewContentListener"`
}

var settings Settings

func init() {
	cfg, err := boilerplate.RecursiveFindFile("config.json", "./", 30)

	if err != nil {
		panic(err)
	}

	if _, err = boilerplate.ReadConfigByFilePaths([]string{cfg}, &settings); err != nil {
		panic(err)
	}
	if boilerplate.GetCurrentEnvironment() == boilerplate.Ci {
		settings.MasterDb.Db = fmt.Sprintf("ci_%v", boilerplate.GetGenerator().Generate().String())
		settings.ReadonlyDb.Db = settings.MasterDb.Db
	}

	CDN_BASE = settings.CdnBase

	// if boilerplate.GetCurrentEnvironment() != boilerplate.Ci && boilerplate.GetCurrentEnvironment() != boilerplate.Local {
	// 	cfgService = application.NewConfigurator[AppConfig]().
	// 		WithRetriever(application.NewHttpRetriever(fmt.Sprintf("%s/internal/json", settings.Wrappers.Configurator.ApiUrl))).
	// 		WithMigrator(application.NewHttpMigrator(fmt.Sprintf("%s/internal/json/migrator", settings.Wrappers.Configurator.ApiUrl)), GetConfigsMigration()).
	// 		MustInit()
	// }
}

func GetConfig() Settings {
	return settings
}

type JobberConfig struct {
	DefaultQueue  string `json:"DefaultQueue"`
	ResultExpire  int    `json:"ResultExpire"`
	Broker        string `json:"Broker"`
	ResultBackend string `json:"ResultBackend"`
	Lock          string `json:"Lock"`
	Concurrency   int    `json:"Concurrency"`
	Tls           bool   `json:"Tls"`
}

func GetJobber(cred *JobberConfig) (*machinery.Server, error) {
	cnf := &config.Config{
		DefaultQueue:    cred.DefaultQueue,
		ResultsExpireIn: cred.ResultExpire,
		Broker:          cred.Broker,
		ResultBackend:   cred.ResultBackend,
		Lock:            cred.Lock,
		NoUnixSignals:   true,
		Redis: &config.RedisConfig{
			MaxIdle:                3,
			IdleTimeout:            240,
			ReadTimeout:            15,
			WriteTimeout:           15,
			ConnectTimeout:         15,
			NormalTasksPollPeriod:  1000,
			DelayedTasksPollPeriod: 500,
		},
	}

	if cred.Tls {
		cnf.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	server, err := boilerplate.NewServer(cnf)

	if err != nil {
		return nil, err
	}

	return server, nil
}
