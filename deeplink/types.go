package deeplink

import (
	"time"

	"github.com/digitalmonsters/go-common/eventsourcing"
)

type IService interface {
	GetVideoShareLink(contentId int64, contentType eventsourcing.ContentType, userId int64, referralCode string) (string, error)
	GetPreviewShareLink(contentId int64, contentType eventsourcing.ContentType, uri string, userId int64, referralCode string) (string, error)
	GetPetVideoShareLink(contentId int64, contentType eventsourcing.ContentType, userId int64, referralCode string, petType int64, petId int64, petName string) (string, error)
	GetVideoShareLinkWithMeta(contentId int64, contentType eventsourcing.ContentType, userId int64, referralCode string, title string, description string, previewThumbnail string) (string, error)
	GenerateBranchDeeplinkWithMeta(link string, title string, description string, previewThumbnail string) (string, error)
}

type firebaseCreateDeeplinkRequest struct {
	DynamicLinkInfo dynamicLinkInfo `json:"dynamicLinkInfo"`
}

type firebaseCreateDeeplinkRequestWithMeta struct {
	DynamicLinkInfo dynamicLinkInfo `json:"dynamicLinkInfo"`
}

type socialnMetaTagInfo struct {
	SocialTitle       string `json:"socialTitle"`
	SocialDescription string `json:"socialDescription"`
	SocialImageLink   string `json:"socialImageLink"`
}

type dynamicLinkInfo struct {
	DomainUriPrefix string             `json:"domainUriPrefix"`
	Link            string             `json:"link"`
	AndroidInfo     androidInfo        `json:"androidInfo"`
	IosInfo         iosInfo            `json:"iosInfo"`
	SocialMetaTag   socialnMetaTagInfo `json:"socialMetaTagInfo"`
}

type androidInfo struct {
	AndroidPackageName string `json:"androidPackageName"`
}

type BranchConfigType struct {
	BranchKey    string `json:"BranchKey"`
	BranchSecret string `json:"BranchSecret"`
}

type iosInfo struct {
	IosBundleId   string `json:"iosBundleId"`
	IosAppStoreId string `json:"iosAppStoreId"`
}

type firebaseCreateDeeplinkResponse struct {
	ShortLink string `json:"shortLink"`
}

type Config struct {
	URI                string           `json:"URI"`
	DomainURIPrefix    string           `json:"DomainURIPrefix"`
	AndroidPackageName string           `json:"AndroidPackageName"`
	IOSBundleId        string           `json:"IOSBundleId"`
	IOSAppStoreId      string           `json:"IOSAppStoreId"`
	Key                string           `json:"Key"`
	BranchConfig       BranchConfigType `json:"BranchConfig"`
	Provider           string           `json:"Provider"`
}

type BranchLinkData struct {
	BranchKey string        `json:"branch_key"`
	Channel   string        `json:"channel"`
	Feature   string        `json:"feature"`
	Campaign  string        `json:"campaign"`
	Stage     string        `json:"stage"`
	Tags      []string      `json:"tags"`
	Type      int           `json:"type"`
	Data      LinkData      `json:"data"`
	Alias     string        `json:"alias"`
	Duration  time.Duration `json:"duration"`
	Analytics AnalyticsData `json:"analytics"`
}

type LinkData struct {
	FallbackURL          string `json:"$fallback_url"`
	FallbackURLXX        string `json:"$fallback_url_xx"`
	DesktopURL           string `json:"$desktop_url"`
	IOSURL               string `json:"$ios_url"`
	IOSURLXX             string `json:"$ios_url_xx"`
	IPadURL              string `json:"$ipad_url"`
	AndroidURL           string `json:"$android_url"`
	AndroidURLXX         string `json:"$android_url_xx"`
	SamsungURL           string `json:"$samsung_url"`
	HuaweiURL            string `json:"$huawei_url"`
	WindowsPhoneURL      string `json:"$windows_phone_url"`
	BlackBerryURL        string `json:"$blackberry_url"`
	FireURL              string `json:"$fire_url"`
	IOSWeChatURL         string `json:"$ios_wechat_url"`
	AndroidWeChatURL     string `json:"$android_wechat_url"`
	WebOnly              bool   `json:"$web_only"`
	DesktopWebOnly       bool   `json:"$desktop_web_only"`
	MobileWebOnly        bool   `json:"$mobile_web_only"`
	AfterClickURL        string `json:"$after_click_url"`
	AfterClickDesktopURL string `json:"$afterclick_desktop_url"`
	CanonicalURL         string `json:"$canonical_url"`
	DeeplinkPath         string `json:"$deeplink_path"`
	OGTitle              string `json:"$og_title"`
	OGDescription        string `json:"$og_description"`
	OGImageURL           string `json:"$og_image_url"`
}

type AnalyticsData struct {
	Channel                    string   `json:"~channel"`
	Feature                    string   `json:"~feature"`
	Campaign                   string   `json:"~campaign"`
	CampaignID                 string   `json:"~campaign_id"`
	CustomerCampaign           string   `json:"~customer_campaign"`
	Stage                      string   `json:"~stage"`
	Tags                       []string `json:"~tags"`
	SecondaryPublisher         string   `json:"~secondary_publisher"`
	CustomerSecondaryPublisher string   `json:"~customer_secondary_publisher"`
	CreativeName               string   `json:"~creative_name"`
	CreativeID                 string   `json:"~creative_id"`
	AdSetName                  string   `json:"~ad_set_name"`
	AdSetID                    string   `json:"~ad_set_id"`
	CustomerAdSetName          string   `json:"~customer_ad_set_name"`
	AdName                     string   `json:"~ad_name"`
	AdID                       string   `json:"~ad_id"`
	CustomerAdName             string   `json:"~customer_ad_name"`
	Keyword                    string   `json:"~keyword"`
	KeywordID                  string   `json:"~keyword_id"`
	CustomerKeyword            string   `json:"~customer_keyword"`
	Placement                  string   `json:"~placement"`
	PlacementID                string   `json:"~placement_id"`
	CustomerPlacement          string   `json:"~customer_placement"`
	SubSiteName                string   `json:"~sub_site_name"`
	CustomerSubSiteName        string   `json:"~customer_sub_site_name"`
}

// BranchLinkRequest represents the structure of the JSON request to Branch API
type BranchLinkRequest struct {
	BranchKey string            `json:"branch_key"`      // Your Branch key for authentication
	Campaign  string            `json:"campaign"`        // Campaign name for the link
	Channel   string            `json:"channel"`         // Channel through which the link is shared
	Feature   string            `json:"feature"`         // Feature associated with the link
	Stage     string            `json:"stage"`           // Stage in the user journey
	Tags      []string          `json:"tags"`            // Tags for categorization
	Alias     string            `json:"alias,omitempty"` // Custom alias for the link
	Data      map[string]string `json:"data"`            // Custom data associated with the link
}

// BranchLinkResponse represents the structure of the JSON response from Branch API
type BranchLinkResponse struct {
	URL string `json:"url"` // The generated dynamic link
}
