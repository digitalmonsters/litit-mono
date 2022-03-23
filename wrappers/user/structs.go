package user

import (
	"github.com/digitalmonsters/go-common/rpc"
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
	UserId                     int64             `json:"user_id"`
	Avatar                     null.String       `json:"avatar"`
	Username                   string            `json:"username"`
	Firstname                  string            `json:"firstname"`
	Lastname                   string            `json:"lastname"`
	Verified                   bool              `json:"verified"`
	Guest                      bool              `json:"guest"`
	EnableAgeRestrictedContent bool              `json:"enable_age_restricted_content"`
	IsTipEnabled               bool              `json:"is_tip_enabled"`
	NamePrivacyStatus          NamePrivacyStatus `json:"name_privacy_status"`
}

type NamePrivacyStatus int

const (
	NamePrivacyStatusVisible         NamePrivacyStatus = 0
	NamePrivacyStatusFirstNameHidden NamePrivacyStatus = 1
	NamePrivacyStatusLastNameHidden  NamePrivacyStatus = 2
	NamePrivacyStatusAllHidden       NamePrivacyStatus = 3
)

type GetUsersRequest struct {
	UserIds []int64 `json:"user_ids"`
}

type GetProfileBulkRequest struct {
	CurrentUserId int64   `json:"request_user_id"`
	UserIds       []int64 `json:"user_ids"`
}

type GetProfileBulkResponseChan struct {
	Error *rpc.RpcError
	Items map[int64]UserDetailRecord `json:"data"`
}

type UserDetailRecord struct {
	Id                int64             `json:"id"`
	Username          null.String       `json:"username"`
	Firstname         string            `json:"firstname"`
	Lastname          string            `json:"lastname"`
	CountryCode       string            `json:"country_code"`
	Gender            null.String       `json:"gender"`
	Following         int               `json:"following"`
	Followers         int               `json:"followers"`
	VideosCount       int               `json:"videos_count"`
	IsFollowing       bool              `json:"is_following"`
	IsFollower        bool              `json:"is_follower"`
	Privacy           UerPrivacy        `json:"privacy"`
	Profile           UserProfile       `json:"profile"`
	CountryName       string            `json:"country_name"`
	Avatar            null.String       `json:"avatar"`
	IsTipEnabled      bool              `json:"is_tip_enabled"`
	Guest             bool              `json:"guest"`
	NamePrivacyStatus NamePrivacyStatus `json:"name_privacy_status"`
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
