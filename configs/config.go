package configs

import (
	_ "embed"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/kelseyhightower/envconfig"
)

type DbConfig struct {
	Host     string `json:"Host" default:"localhost"`
	Port     int    `json:"Port" default:"5432"`
	Db       string `json:"Db" default:"sizzle_test"`
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

type Settings struct {
	HttpPort   int      `json:"HttpPort"`
	AuthApiUrl string   `json:"AuthApiUrl"`
	Db         DbConfig `json:"Db"`
}

var settings Settings

func init() {
	if err := envconfig.Process("", &settings); err != nil {
		panic(err)
	}
}

func GetConfig() Settings {
	return settings
}
