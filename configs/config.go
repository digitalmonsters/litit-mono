package configs

import (
	"fmt"
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/boilerplate"
)

type Settings struct {
	HttpPort        int                  `json:"HttpPort"`
	Wrappers        boilerplate.Wrappers `json:"Wrappers"`
	MasterDb        boilerplate.DbConfig `json:"MasterDb"`
	ReadonlyDb      boilerplate.DbConfig `json:"ReadonlyDb"`
	PrivateHttpPort int                  `json:"PrivateHttpPort"`
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
	if boilerplate.GetCurrentEnvironment() != boilerplate.Ci && boilerplate.GetCurrentEnvironment() != boilerplate.Local {
		cfgService = application.NewConfigurator[AppConfig]().
			WithRetriever(application.NewHttpRetriever(application.HttpRetrieverDefaultUrl)).
			WithMigrator(application.NewHttpMigrator(application.HttpMigratorDefaultUrl), GetConfigsMigration()).
			MustInit()
	}
}

func GetConfig() Settings {
	return settings
}
