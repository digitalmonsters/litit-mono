package configs

import (
	_ "embed"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/boilerplate_testing"
)

type NotifierConfig struct {
	KafkaTopic     boilerplate.KafkaTopicConfig `json:"KafkaTopic"`
	PollTimeMs     int                          `json:"PollTimeMs"`
	WorkerPoolSize int                          `json:"WorkerPoolSize"`
}

type Settings struct {
	HttpPort                             int                                    `json:"HttpPort"`
	PrivateHttpPort                      int                                    `json:"PrivateHttpPort"`
	Wrappers                             boilerplate.Wrappers                   `json:"Wrappers"`
	Db                                   boilerplate.DbConfig                   `json:"Db"`
	KafkaWriter                          *boilerplate.KafkaWriterConfiguration  `json:"KafkaWriter"`
	NotifierCommentConfig                *NotifierConfig                        `json:"NotifierCommentConfig"`
	NotifierContentCommentsCounterConfig *NotifierConfig                        `json:"NotifierContentCommentsCounterConfig"`
	NotifierUserCommentsCounterConfig    *NotifierConfig                        `json:"NotifierUserCommentsCounterConfig"`
	NotifierVoteConfig                   *NotifierConfig                        `json:"NotifierVoteConfig"`
	UserListener                         boilerplate.KafkaListenerConfiguration `json:"UserListener"`
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
		settings.Db.Db = fmt.Sprintf("ci_%v", boilerplate.GetGenerator().Generate().String())

		if err := boilerplate_testing.EnsurePostgresDbExists(settings.Db); err != nil {
			panic(err)
		}
	}

}

func GetConfig() Settings {
	return settings
}
