package configs

import (
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/rs/zerolog/log"
	"os"
)

type Settings struct {
	HttpPort               int                                  `json:"HttpPort"`
	Wrappers               boilerplate.Wrappers                 `json:"Wrappers"`
	MasterDb               boilerplate.DbConfig                 `json:"MasterDb"`
	ReadonlyDb             boilerplate.DbConfig                 `json:"ReadonlyDb"`
	SoundStripe            *SoundStripeConfig                   `json:"SoundStripe"`
	S3                     boilerplate.S3Config                 `json:"S3"`
	PrivateHttpPort        int                                  `json:"PrivateHttpPort"`
	Creators               CreatorsConfig                       `json:"Creators"`
	KafkaWriter            boilerplate.KafkaWriterConfiguration `json:"KafkaWriter"`
	NotifierCreatorsConfig NotifierConfig                       `json:"NotifierCreatorsConfig"`
}

type NotifierConfig struct {
	KafkaTopic boilerplate.KafkaTopicConfig `json:"KafkaTopic"`
	PollTimeMs int                          `json:"PollTimeMs"`
}

type CreatorsConfig struct {
	MaxThresholdHours int `json:"MaxThresholdHours"`
}

type SoundStripeConfig struct {
	ApiUrl     string `json:"ApiUrl"`
	ApiToken   string `json:"ApiToken"`
	MaxWorkers int    `json:"MaxWorkers"`
	MaxTimeout int    `json:"MaxTimeout"`
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
		settings.MasterDb.Db = fmt.Sprintf("ci_%v", int64(os.Getpid()))
		log.Info().Msg(fmt.Sprintf("ci db name generated: %v", settings.MasterDb.Db))
		settings.ReadonlyDb.Db = settings.MasterDb.Db
	}

}

func GetConfig() Settings {
	return settings
}
