package boilerplate

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/skynet2/go-config"
	"github.com/skynet2/go-config/source"
	"github.com/skynet2/go-config/source/env"
	"github.com/skynet2/go-config/source/file"
)

type Environment int32

const (
	Local   Environment = 0
	Dev     Environment = 1
	Qa      Environment = 2
	Staging Environment = 3
	Prod    Environment = 4
	Ci      Environment = 5
	Uat     Environment = 6
)

func (e Environment) ToString() string {
	switch e {
	case Local:
		return "local"
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
	case Uat:
		return "uat"
	default:
		return "unk"
	}
}

type KafkaTopicConfig struct {
	Name              string `json:"Name"`
	NumPartitions     int    `json:"NumPartitions"`
	ReplicationFactor int    `json:"ReplicationFactor"`
	RetentionMs       int    `json:"RetentionMs"`
}

type Wrappers struct {
	Auth                   WrapperConfig `json:"Auth"`
	Content                WrapperConfig `json:"Content"`
	Likes                  WrapperConfig `json:"Likes"`
	Follows                WrapperConfig `json:"Follows"`
	Views                  WrapperConfig `json:"Views"`
	UserCategories         WrapperConfig `json:"UserCategories"`
	UserHashtags           WrapperConfig `json:"UserHashtags"`
	UserDislikes           WrapperConfig `json:"UserDislikes"`
	UserBlock              WrapperConfig `json:"UserBlock"`
	PointsCount            WrapperConfig `json:"PointsCount"`
	AuthGo                 WrapperConfig `json:"AuthGo"`
	NotificationGateway    WrapperConfig `json:"NotificationGateway"`
	NotificationHandler    WrapperConfig `json:"NotificationHandler"`
	UserGo                 WrapperConfig `json:"UserGo"`
	IndecentContentChecker WrapperConfig `json:"IndecentContentChecker"`
	BaseApi                WrapperConfig `json:"BaseApi"`
	GoTokenomics           WrapperConfig `json:"GoTokenomics"`
	SolanaApiGate          WrapperConfig `json:"SolanaApiGate"`
	AdminWs                WrapperConfig `json:"AdminWs"`
	Comment                WrapperConfig `json:"Comment"`
	TesseractOcrApi        WrapperConfig `json:"TesseractOcrApi"`
	Configurator           WrapperConfig `json:"Configurator"`
	ContentUploader        WrapperConfig `json:"ContentUploader"`
	Music                  WrapperConfig `json:"Music"`
	AdsManager             WrapperConfig `json:"AdsManager"`
	BscGateWay             WrapperConfig `json:"BscGateWay"`
}

type WrapperConfig struct {
	ApiUrl     string `json:"ApiUrl"`
	TimeoutSec int    `json:"TimeoutSec"`

	PushPublisher  KafkaBatchWriterV2Configuration `json:"PushPublisher"`
	EmailPublisher KafkaBatchWriterV2Configuration `json:"EmailPublisher"`
}

type ApmConfig struct {
	LogLevel    string `json:"LogLevel"`
	ServiceName string `json:"ServiceName"`
	ServerUrls  string `json:"ServerUrls"`
}

type DbConfig struct {
	Host                     string `json:"Host"`
	Port                     int    `json:"Port"`
	Db                       string `json:"Db"`
	User                     string `json:"User"`
	Password                 string `json:"Password"`
	MaxIdleConnections       int    `json:"MaxIdleConnections"`
	MaxConnectionLifetimeSec int    `json:"MaxConnectionLifetimeSec"`
	MaxOpenConnections       int    `json:"MaxOpenConnections"`
	MaxConnectionIdleSec     int    `json:"MaxConnectionIdleSec"`
}

type RedisConfig struct {
	Host     string `json:"Host"`
	Port     int    `json:"Port"`
	Db       int    `json:"Db"`
	Password string `json:"Password"`
}

type S3Config struct {
	CdnUrl       string `json:"CdnUrl"`
	CdnDirectory string `json:"CdnDirectory"`
	Bucket       string `json:"Bucket"`
	Region       string `json:"Region"`
}

type AzureBlobConfig struct {
	UseCliAuth         bool   `json:"UseCliAuth"`
	StorageAccountName string `json:"AccountName"`
	StorageAccountKey  string `json:"AccountKey"`
	Container          string `json:"Container"`
	CdnUrl             string `json:"CdnUrl"`
	CdnDirectory       string `json:"CdnDirectory"`
}

type KafkaListenerConfiguration struct {
	Hosts                           string     `json:"Hosts"`
	Topic                           string     `json:"Topic"`
	GroupId                         string     `json:"GroupId"`
	KafkaAuth                       *KafkaAuth `json:"KafkaAuth"`
	MinBytes                        int        `json:"MinBytes"`
	MaxBytes                        int        `json:"MaxBytes"`
	Tls                             bool       `json:"Tls"`
	MaxBackOffTimeMilliseconds      int        `json:"MaxBackOffTimeMilliseconds"`
	BackOffTimeIntervalMilliseconds int        `json:"BackOffTimeIntervalMilliseconds"`
}

type KafkaBatchListenerConfiguration struct {
	KafkaListenerConfiguration
	MaxDuration  int `json:"MaxDuration"`
	MaxBatchSize int `json:"MaxBatchSize"`
}

type KafkaWriterConfiguration struct {
	Hosts     string    `json:"Hosts"`
	KafkaAuth KafkaAuth `json:"KafkaAuth"`
	Tls       bool      `json:"Tls"`
}

type KafkaBatchWriterV2Configuration struct {
	Hosts                           string           `json:"Hosts"`
	KafkaAuth                       KafkaAuth        `json:"KafkaAuth"`
	Tls                             bool             `json:"Tls"`
	MaxRetryCount                   int              `json:"MaxRetryCount"`
	FlushTimeMilliseconds           int              `json:"FlushTimeMilliseconds"`
	FlushAtSize                     int              `json:"FlushAtSize"`
	BackOffTimeMaxMilliseconds      int              `json:"BackOffTimeMaxMilliseconds"`
	BackOffTimeIntervalMilliseconds int              `json:"BackOffTimeIntervalMilliseconds"`
	Topic                           KafkaTopicConfig `json:"Topic"`
}

type KafkaAuth struct {
	Type     string `json:"Type"`
	User     string `json:"User"`
	Password string `json:"Password"`
}

func InLocal() bool {
	return GetCurrentEnvironment() == Local
}

func GetCurrentEnvironment() Environment {
	val := os.Getenv("ENVIRONMENT")
	val = strings.ToLower(val)
	switch val {
	case "dev":
		return Dev
	case "qa":
		return Qa
	case "stg":
		return Staging
	case "prod":
		return Prod
	case "ci":
		return Ci
	default:
		// TODO: default environment should be production
		return Local
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
	case Local:
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

	sources = append(sources, env.NewSource(input))

	if err := conf.Load(sources...); err != nil {
		return nil, errors.WithStack(err)
	}

	tempConfig := map[string]interface{}{}

	if err = json.Unmarshal(conf.Get().Bytes(), &tempConfig); err != nil {
		return nil, errors.WithStack(err)
	}

	d, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           input,
		TagName:          "json",
		Squash:           true,
	})

	if err != nil {
		return nil, errors.WithStack(err)
	}

	if err = d.Decode(tempConfig); err != nil {
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
