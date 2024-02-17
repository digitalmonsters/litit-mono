package eventsourcing

import (
	"fmt"
	"time"

	"github.com/digitalmonsters/go-common/translation"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/shopspring/decimal"
	"gopkg.in/guregu/null.v4"
)

type UserEvent struct {
	UserId                 int64                     `json:"user_id"`
	Deleted                bool                      `json:"deleted"`
	CreatedAt              time.Time                 `json:"created_at"`
	UpdatedAt              time.Time                 `json:"updated_at"`
	Avatar                 null.String               `json:"avatar"`
	Username               null.String               `json:"username"`
	Email                  null.String               `json:"email"`
	Firstname              null.String               `json:"firstname"`
	Lastname               null.String               `json:"lastname"`
	Birthdate              null.Time                 `json:"birthdate"`
	AllowNotifications     bool                      `json:"allow_notifications"`
	Newsletter             bool                      `json:"newsletter"`
	CountryCode            null.String               `json:"country_code"`
	HowManyFollowers       null.String               `json:"how_many_followers"`
	IsInfluencer           bool                      `json:"is_influencer"`
	Verified               bool                      `json:"verified"`
	Gender                 null.String               `json:"gender"`
	ReferredById           null.Int                  `json:"referred_by_id"`
	ReferredByType         null.String               `json:"referred_by_type"`
	Followers              int                       `json:"followers"`
	Following              int                       `json:"following"`
	Likes                  int                       `json:"likes"`
	Uploads                int                       `json:"uploads"`
	Views                  int                       `json:"views"`
	Shares                 int                       `json:"shares"`
	Comments               int                       `json:"comments"`
	TotalPoints            decimal.Decimal           `json:"total_points"`
	CurrentPoints          decimal.Decimal           `json:"current_points"`
	CollectedPoints        decimal.Decimal           `json:"collected_points"`
	VaultPoints            decimal.Decimal           `json:"vault_points"`
	AllTimeVaultPoints     decimal.Decimal           `json:"all_time_vault_points"`
	Phone                  null.String               `json:"phone"`
	GoogleUid              null.String               `json:"google_uid"`
	VideosCount            int                       `json:"videos_count"`
	Admin                  bool                      `json:"admin"`
	SuperAdmin             bool                      `json:"super_admin"`
	AreAllVisitorsUnlocked bool                      `json:"are_all_visitors_unlocked"`
	TiktokAvatarKey        null.String               `json:"tiktok_avatar_key"`
	SegmentId              null.Int                  `json:"segment_id"`
	ZammadId               null.Int                  `json:"zammad_id"`
	CreatorStatus          user_go.CreatorStatus     `json:"creator_status"`
	Tags                   null.Int                  `json:"tags"`
	DeviceId               null.String               `json:"device_id"`
	Guest                  bool                      `json:"guest"`
	AppleUid               null.String               `json:"apple_uid"`
	KycStatus              KycStatusType             `json:"kyc_status"`
	KycReason              KycReason                 `json:"kyc_reason"`
	NamePrivacyStatus      user_go.NamePrivacyStatus `json:"name_privacy_status"`
	AdDisabled             bool                      `json:"ad_disabled"`
	GrandReferrer          bool                      `json:"grand_referrer"`
	BannedTill             null.Time                 `json:"banned_till"`
	EmailMarketing         null.String               `json:"email_marketing"`
	EmailMarketingVerified bool                      `json:"email_marketing_verified"`
	Language               translation.Language      `json:"language"`
	Timezone               string                    `json:"timezone"`
	SpotsUploadBanned      bool                      `json:"spots_upload_banned"`
	DogsUploadBanned       bool                      `json:"dogs_upload_banned"`
	PetsUploadBanned       bool                      `json:"pets_upload_banned"`
	BaseChangeEvent
}

const (
	DeleteModeSoft = "soft"
	DeleteModeHard = "hard"
)

func (c UserEvent) GetPublishKey() string {
	return fmt.Sprint(c.UserId)
}

type SocialSubsEvent struct {
	UserId int64 `json:"user_id"`
}

func (e SocialSubsEvent) GetPublishKey() string {
	return fmt.Sprint(e.UserId)
}
