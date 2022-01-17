package configs

import "github.com/digitalmonsters/go-common/boilerplate"

type Settings struct {
	HttpPort    int                  `json:"HttpPort"`
	Wrappers    boilerplate.Wrappers `json:"Wrappers"`
	MasterDb    boilerplate.DbConfig `json:"MasterDb"`
	ReadonlyDb  boilerplate.DbConfig `json:"ReadonlyDb"`
	SoundStripe *SoundStripeConfig   `json:"SoundStripe"`
	PrivateHttpPort int                  `json:"PrivateHttpPort"`
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
}

func GetConfig() Settings {
	return settings
}
