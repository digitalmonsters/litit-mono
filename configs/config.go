package configs

import (
	_ "embed"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/boilerplate_testing"
)

type DbConfig struct {
	Host     string `json:"Host" default:"localhost"`
	Port     int    `json:"Port" default:"5432"`
	Db       string `json:"Db" default:"base_api"`
	User     string `json:"User" default:"postgres"`
	Password string `json:"Password" default:"qwerty"`
}

func (d DbConfig) ToBoilerplate() boilerplate.DbConfig {
	return boilerplate.DbConfig{
		Host:     d.Host,
		Port:     d.Port,
		Db:       d.Db,
		User:     d.User,
		Password: d.Password,
	}
}

type NotifierConfig struct {
	KafkaTopic     string `json:"KafkaTopic"`
	PollTimeMs     int    `json:"PollTimeMs"`
	WorkerPoolSize int    `json:"WorkerPoolSize"`
}

type Settings struct {
	HttpPort                             int                                   `json:"HttpPort"`
	PrivateHttpPort                      int                                   `json:"PrivateHttpPort"`
	Wrappers                             boilerplate.Wrappers                  `json:"Wrappers"`
	Db                                   DbConfig                              `json:"Db"`
	KafkaWriter                          *boilerplate.KafkaWriterConfiguration `json:"KafkaWriter"`
	NotifierCommentConfig                *NotifierConfig                       `json:"NotifierCommentConfig"`
	NotifierContentCommentsCounterConfig *NotifierConfig                       `json:"NotifierContentCommentsCounterConfig"`
	NotifierUserCommentsCounterConfig    *NotifierConfig                       `json:"NotifierUserCommentsCounterConfig"`
	NotifierVoteConfig                   *NotifierConfig                       `json:"NotifierVoteConfig"`
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

	if boilerplate.GetCurrentEnvironment() == boilerplate.Ci{
		settings.Db.Db = fmt.Sprintf("ci_%v", boilerplate.GetGenerator().Generate().String())

		if err := boilerplate_testing.EnsurePostgresDbExists(settings.Db.ToBoilerplate()); err != nil {
			panic(err)
		}
	}

}

func GetConfig() Settings {
	return settings
}
