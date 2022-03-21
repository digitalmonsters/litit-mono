package user_go

import (
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/shopspring/decimal"
	"gopkg.in/guregu/null.v4"
)

type GetUsersResponseChan struct {
	Error *rpc.RpcError
	Items map[int64]UserRecord `json:"items"`
}

type UsersInternalChan struct {
	Error *rpc.RpcError
	UserDetailRecord
}

type GetUsersDetailsResponseChan struct {
	Error *rpc.RpcError
	Items map[int64]UserDetailRecord `json:"items"`
}

//goland:noinspection GoNameStartsWithPackageName
type UserRecord struct {
	UserId                     int64       `json:"user_id"`
	Avatar                     null.String `json:"avatar"`
	Username                   string      `json:"username"`
	Firstname                  string      `json:"firstname"`
	Lastname                   string      `json:"lastname"`
	Verified                   bool        `json:"verified"`
	Guest                      bool        `json:"guest"`
	EnableAgeRestrictedContent bool        `json:"enable_age_restricted_content"`
	IsTipEnabled               bool        `json:"is_tip_enabled"`
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
	Id           int64           `json:"id"`
	Username     null.String     `json:"username"`
	Firstname    string          `json:"firstname"`
	Lastname     string          `json:"lastname"`
	Birthdate    null.Time       `json:"birthdate"`
	KycStatus    string          `json:"kyc_status"`
	CountryCode  string          `json:"country_code"`
	VaultPoints  decimal.Decimal `json:"vault_points"`
	Gender       null.String     `json:"gender"`
	Following    int             `json:"following"`
	Followers    int             `json:"followers"`
	VideosCount  int             `json:"videos_count"`
	Privacy      UerPrivacy      `json:"privacy"`
	Profile      UserProfile     `json:"profile"`
	CountryName  string          `json:"country_name"`
	Avatar       null.String     `json:"avatar"`
	IsTipEnabled bool            `json:"is_tip_enabled"`
	Guest        bool            `json:"guest"`
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
)

type AuthGuestRequest struct {
	DeviceId string `json:"device_id"`
}

type AuthGuestResponseChan struct {
	Error *rpc.RpcError  `json:"error"`
	Data  *AuthGuestResp `json:"data"`
}

type AuthGuestResp struct {
	UserId       int64  `json:"user_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
