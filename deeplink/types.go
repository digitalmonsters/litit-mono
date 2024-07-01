package deeplink

import "github.com/digitalmonsters/go-common/eventsourcing"

type IService interface {
	GetVideoShareLink(contentId int64, contentType eventsourcing.ContentType, userId int64, referralCode string) (string, error)
	GetPreviewShareLink(contentId int64, contentType eventsourcing.ContentType, uri string, userId int64, referralCode string) (string, error)
	GetPetVideoShareLink(contentId int64, contentType eventsourcing.ContentType, userId int64, referralCode string, petType int64, petId int64, petName string) (string, error)
}

type firebaseCreateDeeplinkRequest struct {
	DynamicLinkInfo dynamicLinkInfo `json:"dynamicLinkInfo"`
}

type dynamicLinkInfo struct {
	DomainUriPrefix string      `json:"domainUriPrefix"`
	Link            string      `json:"link"`
	AndroidInfo     androidInfo `json:"androidInfo"`
	IosInfo         iosInfo     `json:"iosInfo"`
}

type androidInfo struct {
	AndroidPackageName string `json:"androidPackageName"`
}

type iosInfo struct {
	IosBundleId   string `json:"iosBundleId"`
	IosAppStoreId string `json:"iosAppStoreId"`
}

type firebaseCreateDeeplinkResponse struct {
	ShortLink string `json:"shortLink"`
}

type Config struct {
	URI                string `json:"URI"`
	DomainURIPrefix    string `json:"DomainURIPrefix"`
	AndroidPackageName string `json:"AndroidPackageName"`
	IOSBundleId        string `json:"IOSBundleId"`
	IOSAppStoreId      string `json:"IOSAppStoreId"`
	Key                string `json:"Key"`
}
