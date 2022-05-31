package user_go

import (
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/digitalmonsters/go-common/translation"
	"github.com/shopspring/decimal"
	"gopkg.in/guregu/null.v4"
	"time"
)

type CreatorStatus int

const (
	CreatorStatusNone     = CreatorStatus(0)
	CreatorStatusPending  = CreatorStatus(1)
	CreatorStatusRejected = CreatorStatus(2)
	CreatorStatusApproved = CreatorStatus(3)
)

type NamePrivacyStatus int

const (
	NamePrivacyStatusVisible         NamePrivacyStatus = 0
	NamePrivacyStatusFirstNameHidden NamePrivacyStatus = 1
	NamePrivacyStatusLastNameHidden  NamePrivacyStatus = 2
	NamePrivacyStatusAllHidden       NamePrivacyStatus = 3
)

//goland:noinspection GoNameStartsWithPackageName
type UserRecord struct {
	UserId                     int64                `json:"user_id"`
	Avatar                     null.String          `json:"avatar"`
	Username                   string               `json:"username"`
	Firstname                  string               `json:"firstname"`
	Lastname                   string               `json:"lastname"`
	Email                      string               `json:"email"`
	Verified                   bool                 `json:"verified"`
	Guest                      bool                 `json:"guest"`
	BannedTill                 null.Time            `json:"banned_till"`
	EnableAgeRestrictedContent bool                 `json:"enable_age_restricted_content"`
	IsTipEnabled               bool                 `json:"is_tip_enabled"`
	NamePrivacyStatus          NamePrivacyStatus    `json:"name_privacy_status"`
	Tags                       Tag                  `json:"tags"`
	Language                   translation.Language `json:"language"`
	CountryCode                null.String          `json:"country_code"`
	Birthdate                  null.Time            `json:"birthdate"`
}

func (u UserRecord) GetFirstAndLastNameWithPrivacy() (string, string) {
	return getFirstAndLastNameWithPrivacy(u.NamePrivacyStatus, u.Firstname, u.Lastname, u.Username)
}

func (u UserRecord) FormatUserName() string {
	return formatUserName(u.Username)
}

type GetUsersRequest struct {
	UserIds []int64 `json:"user_ids"`
}

type GetProfileBulkRequest struct {
	CurrentUserId int64   `json:"request_user_id"`
	UserIds       []int64 `json:"user_ids"`
}

type GetProfileBulkResponseChan struct {
	Error *rpc.RpcError
	Items map[int64]UserProfileDetailRecord `json:"data"`
}

type UserDetailRecord struct {
	Id                  int64                `json:"id"`
	Username            null.String          `json:"username"`
	Firstname           string               `json:"firstname"`
	Lastname            string               `json:"lastname"`
	Birthdate           null.Time            `json:"birthdate"`
	KycStatus           string               `json:"kyc_status"`
	CountryCode         string               `json:"country_code"`
	VaultPoints         decimal.Decimal      `json:"vault_points"`
	Gender              null.String          `json:"gender"`
	Following           int                  `json:"following"`
	Followers           int                  `json:"followers"`
	VideosCount         int                  `json:"videos_count"`
	Privacy             UerPrivacy           `json:"privacy"`
	Profile             UserProfile          `json:"profile"`
	CountryName         string               `json:"country_name"`
	Avatar              null.String          `json:"avatar"`
	IsTipEnabled        bool                 `json:"is_tip_enabled"`
	Guest               bool                 `json:"guest"`
	BannedTill          null.Time            `json:"banned_till"`
	Deleted             bool                 `json:"deleted"`
	NamePrivacyStatus   NamePrivacyStatus    `json:"name_privacy_status"`
	Email               string               `json:"email"`
	Verified            bool                 `json:"verified"`
	Uploads             int                  `json:"uploads"`
	Views               int                  `json:"views"`
	Shares              int                  `json:"shares"`
	Likes               int                  `json:"likes"`
	Comments            int                  `json:"comments"`
	CreatorStatus       CreatorStatus        `json:"creator_status"`
	CreatorRejectReason null.String          `json:"creator_reject_reason"`
	CreatedAt           time.Time            `json:"created_at"`
	AdDisabled          bool                 `json:"ad_disabled"`
	Influencer          bool                 `json:"influencer"`
	Language            translation.Language `json:"language"`
}

func (u UserDetailRecord) GetFirstAndLastNameWithPrivacy() (string, string) {
	return getFirstAndLastNameWithPrivacy(u.NamePrivacyStatus, u.Firstname, u.Lastname, u.Username.ValueOrZero())
}

type UserProfileDetailRecord struct {
	UserDetailRecord
	IsFollowing bool `json:"is_following"`
	IsFollower  bool `json:"is_follower"`
}

type UerPrivacy struct {
	EnablePublicMemberRating           bool `json:"enable_public_member_rating"`
	EnableProfileComments              bool `json:"enable_profile_comments"`
	RestrictProfileCommentsToFollowers bool `json:"restrict_profile_comments_to_followers"`
	EnableContentComments              bool `json:"enable_content_comments"`
	EnableAgeRestrictedProfile         bool `json:"enable_age_restricted_profile"`
	EnablePublicContentLiked           bool `json:"enable_public_content_liked"`
}

type UserProfile struct {
	Id              int64       `json:"id"`
	Bio             string      `json:"bio"`
	AddressCity     null.String `json:"address_city"`
	SocialReddit    string      `json:"social_reddit"`
	SocialQuora     string      `json:"social_quora"`
	SocialMedium    string      `json:"social_medium"`
	SocialLinkedin  string      `json:"social_linkedin"`
	SocialDiscord   string      `json:"social_discord"`
	SocialTelegram  string      `json:"social_telegram"`
	SocialViber     string      `json:"social_viber"`
	SocialWhatsapp  string      `json:"social_whatsapp"`
	SocialFacebook  string      `json:"social_facebook"`
	SocialInstagram null.String `json:"social_instagram"`
	SocialTwitter   null.String `json:"social_twitter"`
	SocialWebsite   null.String `json:"social_website"`
	SocialYoutube   null.String `json:"social_youtube"`
	SocialTiktok    null.String `json:"social_tiktok"`
	SocialClubhouse null.String `json:"social_clubhouse"`
	SocialTwitch    null.String `json:"social_twitch"`
}

type GetUsersDetailRequest struct {
	UserIds []int64 `json:"user_ids"`
}

type GetUsersActiveThresholdsRequest struct {
	UserIds []int64 `json:"user_ids"`
}

type GetUsersActiveThresholdsResponseChan struct {
	Error *rpc.RpcError
	Items map[int64]ThresholdsStruct `json:"items"`
}

type ThresholdsStruct struct {
	Id            int64               `json:"id"`
	ThresholdType ThresholdType       `json:"threshold_type"`
	EntityType    EntityThresholdType `json:"entity_type"`
	EntityId      int64               `json:"entity_id"`
	Amount        null.Int            `json:"amount"`
}

type ThresholdType int

const (
	DailyThreshold      ThresholdType = 1
	WithdrawalThreshold ThresholdType = 2
)

type EntityThresholdType int

const (
	PersonalThresholdEntity EntityThresholdType = 1
	SegmentThresholdEntity  EntityThresholdType = 2
	SystemThresholdEntity   EntityThresholdType = 3
)

type GetUserIdsFilterByUsernameRequest struct {
	UserIds     []int64 `json:"user_ids"`
	SearchQuery string  `json:"search_query"`
}

type GetUserIdsFilterByUsernameResponseChan struct {
	Error   *rpc.RpcError
	UserIds []int64 `json:"user_ids"`
}

type GetUsersTagsRequest struct {
	UserIds []int64 `json:"user_ids"`
}

type GetUsersTagsResponseChan struct {
	Error *rpc.RpcError
	Items map[int64][]Tag `json:"items"`
}

type Tag int

const (
	JunkActivity              Tag = 1 << 0
	LotsOfInvites             Tag = 1 << 1
	ConstantExceedingOfLimits Tag = 1 << 2
	LargeWalletBalance        Tag = 1 << 3
	SuspiciousUser            Tag = 1 << 4
	Bot                       Tag = 1 << 5
)

type AuthGuestRequest struct {
	DeviceId string `json:"device_id"`
}

type AuthGuestResp struct {
	UserId       int64  `json:"user_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type GetBlockListRequest struct {
	UserIds []int64 `json:"user_ids"`
}

type UserBlockData struct {
	Type      *BlockedUserType `json:"type"`
	IsBlocked bool             `json:"is_blocked"`
}

type BlockedUserType string

const (
	BlockedUser   BlockedUserType = "BLOCKED_BY_USER"
	BlockedByUser BlockedUserType = "BLOCKED_TO_USER"
)

type GetUserBlockRequest struct {
	BlockBy   int64 `json:"block_by"`
	BlockedTo int64 `json:"blocked_to"`
}

type UpdateUserMetaDataRequest struct {
	UserId                 int64                `json:"user_id"`
	Email                  null.String          `json:"email"`
	Firstname              null.String          `json:"firstname"`
	Lastname               null.String          `json:"lastname"`
	Birthdate              null.Time            `json:"birthdate"`
	CountryCode            string               `json:"country_code"`
	Username               null.String          `json:"username"`
	Gender                 null.String          `json:"gender"`
	EmailMarketing         null.String          `json:"email_marketing"`
	EmailMarketingVerified bool                 `json:"email_marketing_verified"`
	Language               translation.Language `json:"language"`
}

type ForceResetUserIdentityWithNewGuestRequest struct {
	DeviceId string `json:"device_id"`
}

type ForceResetUserIdentityWithNewGuestResponse struct {
	NewUserId int64 `json:"new_user_id"`
}

type VerifyUserRequest struct {
	UserId int64 `json:"user_id"`
}

type GetAllActiveBotsResponse struct {
	UserIds []int64 `json:"user_ids"`
}

type GetConfigPropertiesResponseChan struct {
	Items map[string]string `json:"items"`
	Error *rpc.RpcError
}
type GetConfigPropertiesRequest struct {
	Properties []string `json:"properties"`
}

type VerifyEmailMarketingRequest struct {
	UserId                 int64 `json:"user_id"`
	EmailMarketingVerified bool  `json:"email_marketing_verified"`
}

type UpdateEmailMarketingRequest struct {
	UserId                 int64       `json:"user_id"`
	EmailMarketing         null.String `json:"email_marketing"`
	EmailMarketingVerified bool        `json:"email_marketing_verified"`
}

type GenerateDeeplinkRequest struct {
	UrlPath string `json:"url"`
}

type GenerateDeeplinkResponse struct {
	Url string `json:"url"`
}

type CreateExportRequest struct {
	Name string     `json:"name"`
	Type ExportType `json:"type"`
}

type CreateExportResponse struct {
	Id int64 `json:"id"`
}

type ExportType int

const (
	ExportTypeUser    = ExportType(1)
	ExportTypeContent = ExportType(2)
)

type FinalizeExportRequest struct {
	ExportId int64       `json:"export_id"`
	File     null.String `json:"file"`
	Error    error       `json:"error"`
}

type FinalizeExportResponse struct {
	Success bool `json:"success"`
}
