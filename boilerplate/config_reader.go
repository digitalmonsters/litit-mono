package boilerplate

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/ft-t/go-micro-env"
	"github.com/pkg/errors"
	"go-micro.dev/v4/config"
	"go-micro.dev/v4/config/source"
	"go-micro.dev/v4/config/source/file"
	"os"
	"path"
	"strings"
)

type Environment int32

const (
	Dev     Environment = 0
	Qa      Environment = 1
	Staging Environment = 2
	Prod    Environment = 3
	Ci      Environment = 4
)

func (e Environment) ToString() string {
	switch e {
	case Dev:
		return "dev"
	case Qa:
		return "qa"
	case Staging:
		return "stg"
	case Prod:
		return "prod"
	case Ci:
		return "ci"
	default:
		return "unk"
	}
}

type Wrappers struct {
	Auth           WrapperConfig `json:"Auth"`
	Content        WrapperConfig `json:"Content"`
	Likes          WrapperConfig `json:"Likes"`
	Views          WrapperConfig `json:"Views"`
	UserCategories WrapperConfig `json:"UserCategories"`
	UserHashtags   WrapperConfig `json:"UserHashtags"`
	UserInfo       WrapperConfig `json:"UserInfo"`
	UserBlock      WrapperConfig `json:"UserBlock"`
	Categories     WrapperConfig `json:"Categories"`
	Hashtags       WrapperConfig `json:"Hashtags"`
}

type WrapperConfig struct {
	ApiUrl     string `json:"ApiUrl"`
	TimeoutSec int    `json:"TimeoutSec"`
}

type ApmConfig struct {
	LogLevel    string `json:"LogLevel"`
	ServiceName string `json:"ServiceName"`
	ServerUrls  string `json:"ServerUrls"`
}

type DbConfig struct {
	Host     string `json:"Host"`
	Port     int    `json:"Port"`
	Db       string `json:"Db"`
	User     string `json:"User"`
	Password string `json:"Password"`
}

type RedisConfig struct {
	Host     string `json:"Host"`
	Port     int    `json:"Port"`
	Db       int    `json:"Db"`
	Password string `json:"Password"`
}

type KafkaListenerConfiguration struct {
	Hosts            string     `json:"Hosts"`
	Topic            string     `json:"Topic"`
	GroupId          string     `json:"GroupId"`
	KafkaAuth        *KafkaAuth `json:"KafkaAuth"`
	MinBytes         int        `json:"MinBytes"`
	MaxBytes         int        `json:"MaxBytes"`
	MaxBatchSize     int        `json:"MaxBatchSize"`
	ListenerDuration int        `json:"ListenerDuration"`
	Tls              bool       `json:"Tls"`
}

type KafkaWriterConfiguration struct {
	Hosts     string    `json:"Hosts"`
	KafkaAuth KafkaAuth `json:"KafkaAuth"`
	Tls       bool      `json:"Tls"`
}

type KafkaAuth struct {
	Type     string `json:"Type"`
	User     string `json:"User"`
	Password string `json:"Password"`
}

func GetCurrentEnvironment() Environment {
	val := os.Getenv("environment")
	val = strings.ToLower(val)
	switch val {
	case "qa":
		return Qa
	case "stg":
		return Staging
	case "prod":
		return Prod
	case "ci":
		return Ci
	default:
		return Dev
	}
}

type ScyllaConfiguration struct {
	Hosts             string `json:"Hosts"`
	UserName          string `json:"UserName"`
	Password          string `json:"Password"`
	Keyspace          string `json:"Keyspace"`
	Enabled           bool   `json:"Enabled"`
	PageSize          int    `json:"PageSize"`
	NumConns          int    `json:"NumConns"`
	MaxRoutingKeyInfo int    `json:"MaxRoutingKeyInfo"`
	MaxPreparedStmts  int    `json:"MaxPreparedStmts"`
	TimeoutSeconds    int    `json:"TimeoutSeconds"`
}

var configuration interface{}

func ReadConfigFile(input interface{}) (interface{}, error) {
	return ReadConfigByFilePaths([]string{"config.json"}, input)
}

func SplitHostsToSlice(currentString string) []string {
	var results []string

	for _, s := range strings.Split(currentString, ",") {
		if trimmed := strings.TrimSpace(s); len(trimmed) > 0 {
			results = append(results, trimmed)
		}
	}

	return results
}

func ReadConfigByFilePaths(filePath []string, input interface{}) (interface{}, error) {
	if configuration != nil {
		return configuration, nil
	}

	if len(filePath) == 0 {
		return nil, errors.New("no config path specified")
	}

	var options []source.Option

	conf, err := config.NewConfig()

	if err != nil {
		return nil, errors.WithStack(err)
	}

	for _, val := range filePath {
		options = addFile(val, options)
	}

	if len(options) == 0 {
		return nil, errors.New("no configuration provided")
	}

	switch GetCurrentEnvironment() {
	case Dev:
		if devFilePath, err := RecursiveFindFile("config.qwerty.json", "./", 30); err == nil {
			options = addFile(devFilePath, options)
		}
	case Ci:
		if ciFilePath, err := RecursiveFindFile("config.ci.json", "./", 30); err != nil {
			return nil, err
		} else {
			options = addFile(ciFilePath, options)
		}
	}

	var sources []source.Source

	for _, option := range options {
		sources = append(sources, file.NewSource(option))
	}

	sources = append(sources, go_micro_env.NewSource(input))

	if err := conf.Load(sources...); err != nil {
		return nil, errors.WithStack(err)
	}

	if err := conf.Get().Scan(input); err != nil {
		return nil, errors.WithStack(err)
	}

	configuration = input

	return configuration, nil
}

func RecursiveFindFile(fileName string, startDirectory string, maxDepth int) (string, error) {
	if len(startDirectory) == 0 {
		startDirectory = "./"
	}

	var checked []string
	for i := 0; i < maxDepth; i++ {
		pathToCheck := path.Join(startDirectory, fileName)

		if _, err := os.Stat(pathToCheck); err == nil {
			return pathToCheck, nil
		}

		checked = append(checked, pathToCheck)

		if startDirectory == "./" {
			startDirectory = "../"
		} else {
			startDirectory = path.Join("../", startDirectory)
		}
	}

	return "", errors.New(fmt.Sprintf("can not find file %v, checked paths : %v", fileName, spew.Sdump(checked)))
}

func addFile(path string, sources []source.Option) []source.Option {
	if _, err := os.Stat(path); err == nil {
		sources = append(sources, file.WithPath(path))
	}
	return sources
}
