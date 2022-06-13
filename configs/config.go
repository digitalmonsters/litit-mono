package configs

import (
	_ "embed"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
)

var CDN_BASE string

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
	SendingQueueCustomListener     boilerplate.KafkaListenerConfiguration `json:"SendingQueueCustomListener"`
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
	EmailLinks                     EmailLinks                             `json:"EmailLinks"`
	Scylla                         boilerplate.ScyllaConfiguration        `json:"Scylla"`
	S3                             boilerplate.S3Config                   `json:"S3"`
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
