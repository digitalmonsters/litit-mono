package configs

import (
	"crypto/tls"
	"fmt"
	"github.com/RichardKnop/logging"
	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	log2 "github.com/RichardKnop/machinery/v1/log"
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
	Feed                   MusicFeedConfiguration               `json:"Feed"`
	Redis                  RedisConfig                          `json:"Redis"`
	Jobber                 JobberConfig                         `json:"Jobber"`
}

type RedisConfig struct {
	Db   int    `json:"Db"`
	Host string `json:"Host"`
	Tls  bool   `json:"Tls"`
}

type JobberConfig struct {
	Tls           bool   `json:"tls"`
	DefaultQueue  string `json:"DefaultQueue"`
	ResultExpire  int    `json:"ResultExpire"`
	Broker        string `json:"Broker"`
	ResultBackend string `json:"ResultBackend"`
	Lock          string `json:"Lock"`
	Concurrency   int    `json:"Concurrency"`
}

type NotifierConfig struct {
	KafkaTopic boilerplate.KafkaTopicConfig `json:"KafkaTopic"`
	PollTimeMs int                          `json:"PollTimeMs"`
}

type CreatorsConfig struct {
	MaxThresholdHours int              `json:"MaxThresholdHours"`
	Listeners         CreatorListeners `json:"Listeners"`
}

type CreatorListeners struct {
	LikeCounter     CounterListener                        `json:"LikeCounter"`
	LoveCounter     CounterListener                        `json:"LoveCounter"`
	DislikeCounter  CounterListener                        `json:"DislikeCounter"`
	ListenCounter   CounterListener                        `json:"ListenCounter"`
	ListenedMusic   CounterListener                        `json:"ListenedMusic"`
	SharesCounter   CounterListener                        `json:"SharesCounter"`
	CommentsCounter boilerplate.KafkaListenerConfiguration `json:"CommentsCounter"`
}

type SoundStripeConfig struct {
	ApiUrl     string `json:"ApiUrl"`
	ApiToken   string `json:"ApiToken"`
	MaxWorkers int    `json:"MaxWorkers"`
	MaxTimeout int    `json:"MaxTimeout"`
}

type CounterListener struct {
	Kafka          boilerplate.KafkaListenerConfiguration
	MaxDuration    int `json:"MaxDuration"`
	MaxBatchSize   int `json:"MaxBatchSize"`
	WorkerPoolSize int `json:"WorkerPoolSize"`
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

func GetJobber(cred JobberConfig) (*machinery.Server, error) {
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

	iface := logging.New(nil, nil, new(logging.ColouredFormatter))

	log2.DEBUG = iface[logging.DEBUG]
	// INFO ...
	log2.INFO = iface[logging.INFO]
	// WARNING ...
	log2.WARNING = iface[logging.WARNING]
	// ERROR ...
	log2.ERROR = iface[logging.ERROR]
	// FATAL ...
	log2.FATAL = iface[logging.FATAL]

	return server, nil
}
