package configs

import (
	"crypto/tls"
	_ "embed"
	"fmt"

	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	"github.com/digitalmonsters/go-common/boilerplate"
)

var CDN_BASE string

const (
	PushNotificationDeadlineKeyMinutes = 720
	PushNotificationDeadlineMinutes    = 720
	PushNotificationJobCron            = "1 */12 * * *" // should be equal to PushNotificationDeadlineMinutes
)

const (
	PREFIX_CONTENT            = "content"
	PREFIX_NOTIFICATION_IMAGE = "notification_image"
)

type Settings struct {
	HttpPort                       int                                    `json:"HttpPort"`
	PrivateHttpPort                int                                    `json:"PrivateHttpPort"`
	CdnBase                        string                                 `json:"CdnBase"`
	Wrappers                       boilerplate.Wrappers                   `json:"Wrappers"`
	MasterDb                       boilerplate.DbConfig                   `json:"MasterDb"`
	ReadonlyDb                     boilerplate.DbConfig                   `json:"ReadonlyDb"`
	KafkaWriter                    boilerplate.KafkaWriterConfiguration   `json:"KafkaWriter"`
	SendingQueueListener           boilerplate.KafkaListenerConfiguration `json:"SendingQueueListener"`
	CreatorsListener               boilerplate.KafkaListenerConfiguration `json:"CreatorsListener"`
	CommentListener                boilerplate.KafkaListenerConfiguration `json:"CommentListener"`
	VoteListener                   boilerplate.KafkaListenerConfiguration `json:"VoteListener"`
	LikeListener                   boilerplate.KafkaListenerConfiguration `json:"LikeListener"`
	ContentListener                boilerplate.KafkaListenerConfiguration `json:"ContentListener"`
	KysStatusListener              boilerplate.KafkaListenerConfiguration `json:"KysStatusListener"`
	FollowListener                 boilerplate.KafkaListenerConfiguration `json:"FollowListener"`
	TokenomicsNotificationListener boilerplate.KafkaListenerConfiguration `json:"TokenomicsNotificationListener"`
	EmailNotificationListener      boilerplate.KafkaListenerConfiguration `json:"EmailNotificationListener"`
	PushAdminMessageListener       boilerplate.KafkaListenerConfiguration `json:"PushAdminMessageListener"`
	UserDeleteListener             boilerplate.KafkaListenerConfiguration `json:"UserDeleteListener"`
	UserBannedListener             boilerplate.KafkaListenerConfiguration `json:"UserBannedListener"`
	UserUpdateListener             boilerplate.KafkaListenerConfiguration `json:"UserUpdateListener"`
	EmailLinks                     EmailLinks                             `json:"EmailLinks"`
	Scylla                         boilerplate.ScyllaConfiguration        `json:"Scylla"`
	AzureBlob                      boilerplate.AzureBlobConfig            `json:"AzureBlob"`
	Jobber                         JobberConfig                           `json:"Jobber"`
	MusicCreatorListener           boilerplate.KafkaListenerConfiguration `json:"MusicCreatorListener"`

	// Firebase Configuration
	Firebase FirebaseConfig `json:"Firebase"`
}

type FirebaseConfig struct {
	ServiceAccountJSON map[string]interface{} `json:"ServiceAccountJSON"`
}

type EmailLinks struct {
	VerifyHost               string `json:"VerifyHost"`
	VerifyPath               string `json:"VerifyPath"`
	MarketingSite            string `json:"MarketingSite"`
	VerifyEmailMarketingPath string `json:"VerifyEmailMarketingPath"`
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

	CDN_BASE = settings.CdnBase
}

func GetConfig() Settings {
	return settings
}

func UpdateScyllaKeyspaceForCiConfig(keyspace string) {
	settings.Scylla.Keyspace = keyspace
}

type JobberConfig struct {
	DefaultQueue  string `json:"DefaultQueue"`
	ResultExpire  int    `json:"ResultExpire"`
	Broker        string `json:"Broker"`
	ResultBackend string `json:"ResultBackend"`
	Lock          string `json:"Lock"`
	Concurrency   int    `json:"Concurrency"`
	Tls           bool   `json:"Tls"`
}

func GetJobber(cred *JobberConfig) (*machinery.Server, error) {
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

	return server, nil
}

type MachineryTask string

const (
	GeneralPushNotificationTask  MachineryTask = "general:push_notification"
	PeriodicPushNotificationTask MachineryTask = "periodic:push_notification"
	UserPushNotificationTask     MachineryTask = "user:push_notification"
)
