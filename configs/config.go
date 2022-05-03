package configs

import (
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/rs/zerolog/log"
)

type Settings struct {
	HttpPort        int                  `json:"HttpPort"`
	PrivateHttpPort int                  `json:"PrivateHttpPort"`
	MasterDb        boilerplate.DbConfig `json:"MasterDb"`
	ReadonlyDb      boilerplate.DbConfig `json:"ReadonlyDb"`
	Wrappers        boilerplate.Wrappers `json:"Wrappers"`
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
		settings.MasterDb.Db = boilerplate_testing.GetPostgresCiDatabaseName()
		log.Info().Msg(fmt.Sprintf("ci db name generated: %v", settings.MasterDb.Db))
		settings.ReadonlyDb.Db = settings.MasterDb.Db
	}
}

func GetConfig() Settings {
	return settings
}
