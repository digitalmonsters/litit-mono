package configs

import (
	_ "embed"
	"fmt"

	"github.com/digitalmonsters/go-common/boilerplate"
)

type Settings struct {
	HttpPort                   int                                    `json:"HttpPort"`
	PrivateHttpPort            int                                    `json:"PrivateHttpPort"`
	Wrappers                   boilerplate.Wrappers                   `json:"Wrappers"`
	MasterDb                   boilerplate.DbConfig                   `json:"MasterDb"`
	ReadonlyDb                 boilerplate.DbConfig                   `json:"ReadonlyDb"`
	KafkaWriter                boilerplate.KafkaWriterConfiguration   `json:"KafkaWriter"`
	SendingQueueListener       boilerplate.KafkaListenerConfiguration `json:"SendingQueueListener"`
	SendingQueueCustomListener boilerplate.KafkaListenerConfiguration `json:"SendingQueueCustomListener"`
	CreatorsListener           boilerplate.KafkaListenerConfiguration `json:"CreatorsListener"`
	MusicCreatorsListener      boilerplate.KafkaListenerConfiguration `json:"MusicCreatorsListener"`
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

}

func GetConfig() Settings {
	return settings
}
