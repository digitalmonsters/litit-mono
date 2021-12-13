package configs

import (
	_ "embed"
	"github.com/digitalmonsters/go-common/boilerplate"
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
	HttpPort                             int                                   `json:"HttpPort" default:"5215"`
	Wrappers                             boilerplate.Wrappers                  `json:"Wrappers"`
	Db                                   DbConfig                              `json:"Db"`
	KafkaWriter                          *boilerplate.KafkaWriterConfiguration `json:"KafkaWriter"`
	NotifierCommentConfig                *NotifierConfig                       `json:"NotifierCommentConfig"`
	NotifierContentCommentsCounterConfig *NotifierConfig                       `json:"NotifierContentCommentsCounterConfig"`
	NotifierUserCommentsCounterConfig    *NotifierConfig                       `json:"NotifierUserCommentsCounterConfig"`
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
}

func GetConfig() Settings {
	return settings
}
