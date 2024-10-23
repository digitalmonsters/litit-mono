package go_tokenomics

import (
	"fmt"
	"time"

	"github.com/digitalmonsters/go-common/filters"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/shopspring/decimal"
)

type GetUsersTokenomicsInfoRequest struct {
	UserIds []int64          `json:"user_ids"`
	Filters []filters.Filter `json:"filters"`
}

type UserTokenomicsInfo struct {
	TotalPoints        decimal.Decimal `json:"total_points"`
	CurrentPoints      decimal.Decimal `json:"current_points"`
	VaultPoints        decimal.Decimal `json:"vault_points"`
	AllTimeVaultPoints decimal.Decimal `json:"all_time_vault_points"`
	CurrentTokens      decimal.Decimal `json:"current_tokens"`
	CurrentRate        decimal.Decimal `json:"current_rate"`
	WithdrawnTokens    decimal.Decimal `json:"withdrawn_tokens"`
}

type SendMessageResponseChan struct {
	Error *rpc.RpcError `json:"error"`
}

type GetWithdrawalsAmountsByAdminIdsRequest struct {
	AdminIds []int64 `json:"admin_ids"`
}

type GetWithdrawalsAmountsByAdminIdsResponseChan struct {
	Items map[int64]decimal.Decimal
	Error *rpc.RpcError
}

type GetContentEarningsTotalByContentIdsRequest struct {
	ContentIds []int64 `json:"content_ids"`
}

type GetContentEarningsTotalByContentIdsResponseChan struct {
	Items map[int64]decimal.Decimal
	Error *rpc.RpcError
}

type GetTokenomicsStatsByUserIdRequest struct {
	UserIds []int64 `json:"user_ids"`
}

type GetTokenomicsStatsByUserIdResponseChan struct {
	Items map[int64]*UserTokenomicsStats `json:"items"`
	Error *rpc.RpcError
}

type PointsEarnedType int16

const (
	PointsEarnedTypeNone                                      = PointsEarnedType(0)
	PointsEarnedTypeBonusVerify                               = PointsEarnedType(1)  // "extra bonus"
	PointsEarnedTypeBonusVerifyReferrer                       = PointsEarnedType(2)  // "bonus verify referer"
	PointsEarnedTypeTipSent                                   = PointsEarnedType(3)  // "tip sent"
	PointsEarnedTypeTipReceived                               = PointsEarnedType(4)  // "tip received"
	PointsEarnedTypeWithdraw                                  = PointsEarnedType(5)  // "withdraw"
	PointsEarnedTypeDailyTime                                 = PointsEarnedType(6)  // "daily time"
	PointsEarnedTypeUserView                                  = PointsEarnedType(7)  // "user view"
	PointsEarnedTypeSharerView                                = PointsEarnedType(8)  // "sharer view"
	PointsEarnedTypeCreatorView                               = PointsEarnedType(9)  // "creator view"
	PointsEarnedTypeBonusPerformance                          = PointsEarnedType(10) // "bonus performance"
	PointsEarnedTypeDailyFollowers                            = PointsEarnedType(11) // "daily followers"
	PointsEarnedTypeUserShare                                 = PointsEarnedType(12) // "user share"
	PointsEarnedTypeVerifyGrandReferer                        = PointsEarnedType(13) // "bonus verify grand referer"
	PointsEarnedTypeAdmin                                     = PointsEarnedType(14) // "admin"
	PointsEarnedTypeCreatorShare                              = PointsEarnedType(15) // "creator share"
	PointsEarnedTypeCreatorShareGuest                         = PointsEarnedType(16) // "creator share guest"
	PointsEarnedTypeKycPayment                                = PointsEarnedType(17) // "kyc payment"
	PointsEarnedTypeBonusKyc                                  = PointsEarnedType(18) // "bonus kyc"
	PointsEarnedTypeCreatorViewGuest                          = PointsEarnedType(19) // "creator view guest"
	PointsEarnedTypeTapJoy                                    = PointsEarnedType(20) // "tapjoy"
	PointsEarnedTypeWeeklyTime                                = PointsEarnedType(21) // "weekly time"
	PointsEarnedTypeWeeklyFollowers                           = PointsEarnedType(22) // "weekly followers"
	PointsEarnedTypeSharerViewGuest                           = PointsEarnedType(23) // "sharer view guest"
	PointsEarnedTypeTransfer                                  = PointsEarnedType(24) // new type
	PointsEarnedTypeFeatureBought                             = PointsEarnedType(25)
	PointsEarnedTypeApplovin                                  = PointsEarnedType(26) // "applovin"
	PointsEarnedTypeWithdrawRejected                          = PointsEarnedType(27)
	PointsEarnedTypeMegaBonus                                 = PointsEarnedType(28) // "mega bonus"
	PointsEarnedTypeAvatarAdded                               = PointsEarnedType(29) // "avatar added"
	PointsEarnedTypeDescription                               = PointsEarnedType(30) // "adding description"
	PointsEarnedTypeUploadFirstVideo                          = PointsEarnedType(31)
	PointsEarnedTypePointsWriteOff                            = PointsEarnedType(32) // write off money
	PointsEarnedTypeTechnicTransfer                           = PointsEarnedType(33) //technical transfer
	PointsEarnedTypeUploadFirstSpot                           = PointsEarnedType(34)
	PointsEarnedTypeSpotsUserView                             = PointsEarnedType(35)
	PointsEarnedTypeSpotsCreatorView                          = PointsEarnedType(36)
	PointsEarnedTypeSpotsCreatorViewGuest                     = PointsEarnedType(37)
	PointsEarnedTypeSpotsBonusPerformance                     = PointsEarnedType(38)
	PointsEarnedTypeEmailMarketingAdded                       = PointsEarnedType(39) // "email marketing added"
	PointsEarnedTypeIronsource                                = PointsEarnedType(40) // "ironsource"
	PointsEarnedTypeTopSpotDaily                              = PointsEarnedType(41)
	PointsEarnedTypeTopSpotWeekly                             = PointsEarnedType(42)
	PointsEarnedTypeSocialSubsTargetAchieved                  = PointsEarnedType(43)
	PointsEarnedTypeMonthlyTimeMegaBonus                      = PointsEarnedType(44)
	PointsEarnedTypeUploadFirstBioVideo                       = PointsEarnedType(45)
	PointsEarnedTypeAdditionalBonusCreatorVerifyGrandReferrer = PointsEarnedType(46)
	PointsEarnedTypeSocialMediasAdded                         = PointsEarnedType(47)
	PointsEarnedTypePointsWriteOffForAd                       = PointsEarnedType(48) // write off money for ad
	PointsEarnedTypeUserShortListen                           = PointsEarnedType(49)
	PointsEarnedTypeUserCreatorShortListen                    = PointsEarnedType(50)
	PointsEarnedTypeCreatorShortListenGuest                   = PointsEarnedType(51)
	PointsEarnedTypeShortListenBonusPerformance               = PointsEarnedType(52)
	PointsEarnedTypeUserFullListen                            = PointsEarnedType(53)
	PointsEarnedTypeUserCreatorFullListen                     = PointsEarnedType(54)
	PointsEarnedTypeCreatorFullListenGuest                    = PointsEarnedType(55)
	PointsEarnedTypeFullListenBonusPerformance                = PointsEarnedType(56)
	PointsEarnedTypeGoogleAds                                 = PointsEarnedType(57) // google ads
	PointsEarnedTypeSurvey                                    = PointsEarnedType(58) // survey
	PointsEarnedTypeBonusVerifyReferrerWatchedVideo           = PointsEarnedType(59) // "bonus verify referer watched video"
	PointsEarnedTypeBonusReferFriendsWithinWeeklyPeriod       = PointsEarnedType(60) // "bonus refer x friends within the week period"
	PointsEarnedTypeUserDeductPointWeeklyLogin                = PointsEarnedType(61)
	PointsEarnedTypeUserDeductPointWeeklySpendTime            = PointsEarnedType(62)
)

type WithdrawalStatus int16

const (
	WithdrawalStatusNone                         WithdrawalStatus = 0
	WithdrawalStatusPending                      WithdrawalStatus = 1
	WithdrawalStatusApproved                     WithdrawalStatus = 2
	WithdrawalStatusRejected                     WithdrawalStatus = 3
	WithdrawalStatusFailed                       WithdrawalStatus = 4
	WithdrawalStatusPaymentPending               WithdrawalStatus = 5 // [DEPRECATED] not use anymore
	WithdrawalStatusPaid                         WithdrawalStatus = 6 // [DEPRECATED] not use anymore
	WithdrawalStatusPaymentInvestigationRequired WithdrawalStatus = 7 // [DEPRECATED] not use anymore
	WithdrawalStatusInitialClaim                 WithdrawalStatus = 8
	WithdrawalStatusExpired                      WithdrawalStatus = 9
	WithdrawalStatusClaimed                      WithdrawalStatus = 10
)

func (s WithdrawalStatus) ToString() string {
	switch s {
	case WithdrawalStatusPending:
		return "pending"
	case WithdrawalStatusApproved:
		return "approved"
	case WithdrawalStatusRejected:
		return "rejected"
	case WithdrawalStatusFailed:
		return "failed"
	case WithdrawalStatusPaymentPending:
		return "payment pending"
	case WithdrawalStatusPaid:
		return "paid"
	case WithdrawalStatusPaymentInvestigationRequired:
		return "payment investigation required"
	case WithdrawalStatusInitialClaim:
		return "initial claim"
	case WithdrawalStatusExpired:
		return "expired"
	case WithdrawalStatusClaimed:
		return "claimed"
	default:
		return fmt.Sprint(s)
	}
}

type UserTokenomicsStats struct {
	LITITBalance decimal.Decimal `json:"litit_balance"`

	PointsEarnedStats map[PointsEarnedType]UserTokenomicsPointsEarnedStats `json:"points_earned_stats"`
	WithdrawalsStats  map[WithdrawalStatus]UserTokenomicsWithdrawalsStats  `json:"withdrawals_stats"`
}

type UserTokenomicsPointsEarnedStats struct {
	TotalOperationsAmount       decimal.Decimal `json:"total_operations_amount"`        // sum of ALL OPERATIONS ; always increase
	TotalOperationsAmountTokens decimal.Decimal `json:"total_operations_amount_tokens"` // sum of ALL OPERATIONS ; always increase
	TotalOperationsCount        int             `json:"total_operations_count"`
}

type UserTokenomicsWithdrawalsStats struct {
	TotalOperationsCount  int             `json:"total_operations_count"`
	TotalOperationsAmount decimal.Decimal `json:"total_operations_amount"`
}

type GetConfigPropertiesResponseChan struct {
	Items map[string]string `json:"items"`
	Error *rpc.RpcError
}
type GetConfigPropertiesRequest struct {
	Properties []string `json:"properties"`
}

type GetReferralInfoRequest struct {
	ReferrerId  int64   `json:"referrer_id"`
	ReferralIds []int64 `json:"referral_ids"`
}

type GetReferralProgressInfoRequest struct {
	ReferrerId int64 `json:"referrer_id"`
}

type GetMyReferredUsersWatchedVideoInfoRequest struct {
	ReferrerId int64 `json:"referrer_id"`
	Page       int64 `json:"page"`
	Count      int64 `json:"count"`
}

type GetReferralInfoResponse struct {
	TotalCollectedPoints      decimal.Decimal           `json:"total_collected_points"`
	Referrals                 map[int64]decimal.Decimal `json:"referrals"`
	TotalGrandCollectedPoints decimal.Decimal           `json:"total_grand_collected_points"`
}

type ReferralInfo struct {
	CollectedPoints  decimal.Decimal `json:"collected_points"`
	CurrentReferrals int             `json:"current_referrals"`
	TargetReferrals  int64           `json:"target_referrals"`
	PointRate        decimal.Decimal `json:"point_rate"`
	TargetRate       int             `json:"target_rate"`
}
type ReferralGroup struct {
	Type     PointsEarnedType `json:"type"`
	TypeName string           `json:"type_name"`
	Data     ReferralInfo     `json:"data"`
}

type GetReferralProgressInfoResponse struct {
	ListProgress     []ReferralGroup          `json:"list_progress"`
	ReferralUserInfo []ReferredUsersWatchTime `json:"referral_user_info"`
}

type ReferredUsersWatchTime struct {
	Id          int64     `gorm:"primary_key;column:id"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
	UserId      int64     `gorm:"column:user_id"`
	ReferrerId  int64     `gorm:"column:referrer_id"`
	Hours       int32     `gorm:"column:hours"`
	IsBonus     bool      `gorm:"column:is_bonus"`
	PointEarned int32     `gorm:"column:point_earned"`
	TargetHours int32     `gorm:"column:target_hours"`
}

type GetMyReferredUsersWatchedVideoInfoResponse struct {
	ReferredUsersWatchTime []ReferredUsersWatchTimeInfo `json:"referred_users_watch_time"`
	TotalCount             int64                        `json:"total_count"`
}

type DeductVaultPointsForIntroFeedResponse struct {
	Status bool `json:"status"`
}

type DeductVaultPointsForIntroFeedRequest struct {
	UserId int64 `json:"user_id"`
}

type ReferredUsersWatchTimeInfo struct {
	Id          int64     `json:"id"`
	UserId      int64     `json:"user_id"`
	ReferrerId  int64     `json:"referrer_id"`
	Hours       int32     `json:"hours"`
	IsBonus     bool      `json:"is_bonus"`
	PointEarned int32     `json:"point_earned"`
	TargetHours int32     `json:"target_hours"`
	CreatedAt   time.Time `json:"created_at"`
}

type GetActivitiesInfoRequest struct {
	UserId int64 `json:"user_id"`
}

type GetActivitiesInfoResponse struct {
	Items map[int64]UserActivity `json:"items"`
	Error *rpc.RpcError
}

type UserActivity struct {
	AddDescriptionBonus      bool `json:"add_description_bonus"`
	FirstVideoUploaded       bool `json:"first_video_uploaded"`
	FirstBioVideoUploaded    bool `json:"first_bio_video_uploaded"`
	FirstSpotUploaded        bool `json:"first_spot_uploaded"`
	FirstTimeAvatarAdded     bool `json:"first_time_avatar_added"`
	FirstEmailMarketingAdded bool `json:"first_email_marketing_added"`
	FirstXSocialMediaAdded   bool `json:"first_x_social_media_added"`
}

type FilterField string

const (
	TotalPoints        FilterField = "total_points"
	CurrentPoints      FilterField = "current_points"
	VaultPoints        FilterField = "vault_points"
	AllTimeVaultPoints FilterField = "all_time_vault_points"
	CurrentTokens      FilterField = "current_tokens"
	CurrentRate        FilterField = "current_rate"
	WithdrawnTokens    FilterField = "withdrawn_tokens"
)

type CreateBotViewsRequest struct {
	BotViews map[int64][]int64 `json:"bot_views"`
}

type WriteOffUserTokensForAdRequest struct {
	UserId       int64           `json:"user_id"`
	AdCampaignId int64           `json:"ad_campaign_id"`
	Amount       decimal.Decimal `json:"amount"`
}

type PointToVaultType string

const (
	EXTRA_BONUS                 = PointToVaultType("EXTRA_BONUS")
	VERIFIED_REFERRALS          = PointToVaultType("VERIFIED_REFERRALS")
	VIDEOS_SHARED               = PointToVaultType("VIDEOS_SHARED")
	VIDEO_VIEWED                = PointToVaultType("VIDEO_VIEWED")
	MY_VIDEO_VIEWS_GENERATED    = PointToVaultType("MY_VIDEO_VIEWS_GENERATED")
	MY_SPOTS_VIEWS_GENERATED    = PointToVaultType("MY_SPOTS_VIEWS_GENERATED")
	DAILY_TIME                  = PointToVaultType("DAILY_TIME")
	DAILY_FOLLOWERS             = PointToVaultType("DAILY_FOLLOWERS")
	WEEKLY_TIME                 = PointToVaultType("WEEKLY_TIME")
	WEEKLY_FOLLOWERS            = PointToVaultType("WEEKLY_FOLLOWERS")
	TAPJOY                      = PointToVaultType("TAPJOY")
	APPLOVIN                    = PointToVaultType("APPLOVIN")
	GOOGLE_ADS                  = PointToVaultType("GOOGLE_ADS")
	IRONSOURCE                  = PointToVaultType("IRONSOURCE")
	MEGA_BONUS                  = PointToVaultType("MEGA_BONUS")
	FIRST_VIDEO                 = PointToVaultType("FIRST_VIDEO")
	FIRST_SPOT                  = PointToVaultType("FIRST_SPOT")
	FIRST_BIO_VIDEO             = PointToVaultType("FIRST_BIO_VIDEO")
	DESCRIPTION_BONUS           = PointToVaultType("DESCRIPTION_BONUS")
	AVATAR_ADDED                = PointToVaultType("AVATAR_ADDED")
	SECOND_LEVEL_REFERRALS      = PointToVaultType("SECOND_LEVEL_REFERRALS")
	EMAIL_MARKETING_ADDED       = PointToVaultType("EMAIL_MARKETING_ADDED")
	DAILY_TOP_SPOT              = PointToVaultType("DAILY_TOP_SPOT")
	WEEKLY_TOP_SPOT             = PointToVaultType("WEEKLY_TOP_SPOT")
	SOCIAL_SUBS_TARGET_ACHIEVED = PointToVaultType("SOCIAL_SUBS_TARGET_ACHIEVED")
	MONTHLY_TIME_MEGA_BONUS     = PointToVaultType("MONTHLY_TIME_MEGA_BONUS")
	FIRST_SOCIAL_MEDIAS_ADDED   = PointToVaultType("FIRST_SOCIAL_MEDIAS_ADDED")
	SURVEY                      = PointToVaultType("SURVEY")
	FRIENDS_WATCHED_10HRS_VIDEO = PointToVaultType("FRIENDS_WATCHED_10HRS_VIDEO")
	WEEKLY_REFERRALS            = PointToVaultType("WEEKLY_REFERRALS")
	FRIENDS_WATCHED_10ADS       = PointToVaultType("FRIENDS_WATCHED_10ADS")
)

var ALL_FRONTEND_REWARD_TYPES = []PointToVaultType{
	VERIFIED_REFERRALS,
	FRIENDS_WATCHED_10HRS_VIDEO,
	SECOND_LEVEL_REFERRALS,
	VIDEOS_SHARED,
	VIDEO_VIEWED,
	MY_VIDEO_VIEWS_GENERATED,
	MY_SPOTS_VIEWS_GENERATED,
	TAPJOY,
	DAILY_TIME,
	DAILY_FOLLOWERS,
	WEEKLY_TIME,
	WEEKLY_FOLLOWERS,
	WEEKLY_REFERRALS,
	DAILY_TOP_SPOT,
	WEEKLY_TOP_SPOT,
	MONTHLY_TIME_MEGA_BONUS,
	SURVEY,
	FRIENDS_WATCHED_10ADS,
}

type TokenomicsAppConfig struct {
	TOKENOMICS_FEATURE_ONE_PAID_SHARE_PER_LINK_ENABLED bool
	POINT_BONUS_VERIFY                                 decimal.Decimal // ok
	POINT_BONUS_REFERRER_VERIFY                        decimal.Decimal // ok
	POINT_BONUS_GRANDREFERRER_VERIFY                   decimal.Decimal

	TOKEN_TOTAL_SUPPLY decimal.Decimal // important not to change
	POINT_TO_TOKEN     decimal.Decimal // important not to change

	TOKEN_RESERVE_SUPPLY                                   decimal.Decimal // important
	POINT_VIEW_CC_USER                                     decimal.Decimal
	POINT_VIEW_CC_GUEST                                    decimal.Decimal // to remove, not using
	PAID_VIEW_MAX_PER_USER                                 int64
	DAILY_MAX_AMOUNT_OF_PAID_VIEWS                         int64
	PAID_VIEW_THRESHOLD_PERCENT                            decimal.Decimal
	PAID_VIEW_THRESHOLD_TIME                               int64 // in seconds
	PAID_VIEW_TIMEOUT_PER_USER                             int64 // todo check ; in days
	POINT_VIEW_USER                                        decimal.Decimal
	POINT_VIEW_SHARER                                      decimal.Decimal // todo check with igor
	THRESHOLD_FOR_PERFORMANCE_BONUS_FOR_CONTENT_PAID_VIEWS int64           // todo will not work
	TOKENS_FOR_PERFORMANCE_BONUS_OF_CONTENT_PAID_VIEWS     decimal.Decimal // todo will not work
	PAID_SHARE_MAX_PER_USER                                int64
	DAILY_MAX_AMOUNT_OF_PAID_SHARES                        int64
	PAID_SHARE_TIMEOUT_PER_USER                            int64
	PAID_SHARE_VIEW_THRESHOLD_PERCENT                      decimal.Decimal // todo check with igor;
	PAID_SHARE_VIEW_THRESHOLD_TIME                         int64           // todo check with igor;
	POINT_SHARE_USER                                       decimal.Decimal
	POINTS_THRESHOLD_FOR_USER_WITHOUT_KYC                  decimal.Decimal
	POINT_BASE_RATE                                        decimal.Decimal // important
	POINT_SHARE_CC_USER                                    decimal.Decimal
	POINT_SHARE_CC_GUEST                                   decimal.Decimal
	TOKENOMICS_WITHDRAW_LIMIT_BEFORE_VESTING               decimal.Decimal

	VISITORS_UNLOCKED_FEATURE_PRICE decimal.Decimal

	POINT_BONUS_DAILY_FOLLOWERS_TARGET int64
	POINT_BONUS_DAILY_FOLLOWERS        decimal.Decimal

	POINT_BONUS_DAILY_TIME                      decimal.Decimal
	POINT_BONUS_DAILY_TIME_TARGET               int64
	MAX_TIP_POINTS_PER_DAY_FOR_USER_WITHOUT_KYC decimal.Decimal

	POINT_BONUS_WEEKLY_TIME      decimal.Decimal
	POINT_BONUS_WEEKLY_FOLLOWERS decimal.Decimal
	WEEKLY_TIME_TARGET           int64
	WEEKLY_FOLLOWERS_TARGET      int64

	TOKENOMICS_FIRST_X_PAID_VIEWS_COUNT                  int64
	TOKENOMICS_FIRST_X_PAID_VIEWS_AS_CONTENT_OWNER_COUNT int64
	TOKENIMICS_REFERRAL_INCREASE_MULTIPLIER_STEP_1       decimal.Decimal
	TOKENIMICS_REFERRAL_INCREASE_MULTIPLIER_STEP_2       decimal.Decimal
	TOKENOMICS_REFERRAL_EARNED_POINTS_THRESHOLD_STEP_1   decimal.Decimal
	TOKENOMICS_FIRST_X_PAID_VIEWS_COUNT_GENDER_PUSH      int64

	GUEST_TOKENOMICS_FIRST_X_PAID_VIEWS_COUNT int64
	GUEST_TOKENOMICS_FIRST_X_EARNED_POINTS    decimal.Decimal
	GUEST_MAX_EARNED_POINTS_FOR_VIEWS         decimal.Decimal

	MEGA_BONUS_TIME_LIMIT_AFTER_START_MINUTES   int64
	MEGA_BONUS_REFERRALS_TARGET                 int64
	MEGA_BONUS_REFERRALS_POINTS_AMOUNT          decimal.Decimal
	MEGA_BONUS_REFERRALS_VIDEO_URL              string
	MEGA_BONUS_REFERRALS_CONTENT_ID             string
	MEGA_BONUS_MIN_TOTAL_WATCH_TIME_REQUIREMENT int64

	POINT_BONUS_DESCRIPTION              decimal.Decimal
	AVATAR_ADDED_BONUS_POINTS_AMOUNT     decimal.Decimal
	POINT_BONUS_FIRST_VIDEO_UPLOADED     decimal.Decimal
	POINT_BONUS_FIRST_BIO_VIDEO_UPLOADED decimal.Decimal

	GUEST_AFTER_INSTALL_FIRST_PUSH_CONDITION_POINTS        decimal.Decimal
	GUEST_AFTER_INSTALL_SECOND_PUSH_CONDITION_POINTS       decimal.Decimal
	GUEST_AFTER_INSTALL_THIRD_PUSH_CONDITION_POINTS        decimal.Decimal
	GUEST_AFTER_INSTALL_FIRST_PUSH_CONDITION_TIME_MINUTES  int64
	GUEST_AFTER_INSTALL_SECOND_PUSH_CONDITION_TIME_MINUTES int64
	GUEST_AFTER_INSTALL_THIRD_PUSH_CONDITION_TIME_MINUTES  int64

	USER_AFTER_SIGNUP_FIRST_PUSH_CONDITION_CONTENT_ID    int64
	USER_AFTER_SIGNUP_FIRST_PUSH_CONDITION_TIME_MINUTES  int64
	USER_AFTER_SIGNUP_SECOND_PUSH_CONDITION_TIME_MINUTES int64
	USER_AFTER_SIGNUP_THIRD_PUSH_CONDITION_TIME_MINUTES  int64
	USER_AFTER_SIGNUP_FOURTH_PUSH_CONDITION_TIME_HOURS   int64
	USER_AFTER_SIGNUP_FIFTH_PUSH_CONDITION_TIME_HOURS    int64
	KYC_GRACE_PERIOD                                     int64 //days

	AD_POINTS_AMOUNT_REWARD         decimal.Decimal // applovin
	IRONSOURCE_REWARD               decimal.Decimal // ironsource
	POINT_BONUS_FIRST_SPOT_UPLOADED decimal.Decimal

	POINT_SPOTS_VIEW_USER                                        decimal.Decimal
	POINT_SPOTS_VIEW_CC_USER                                     decimal.Decimal
	DAILY_MAX_AMOUNT_OF_PAID_SPOTS_VIEWS                         int64
	PAID_SPOTS_VIEW_THRESHOLD_PERCENT                            decimal.Decimal
	PAID_SPOTS_VIEW_THRESHOLD_TIME                               int64
	THRESHOLD_FOR_PERFORMANCE_BONUS_FOR_CONTENT_PAID_SPOTS_VIEWS decimal.Decimal
	TOKENS_FOR_PERFORMANCE_BONUS_OF_CONTENT_PAID_SPOTS_VIEWS     decimal.Decimal

	WITHDRAWAL_WHITELIST_USER_IDS        string
	WITHDRAWAL_CREATION_INTERVAL_MINUTES int64
	WITHDRAWAL_BLOCKED_COUNTRIES         string
	WITHDRAWAL_FEATURE_ENABLED           bool

	EMAIL_MARKETING_BONUS_POINTS_AMOUNT decimal.Decimal

	DAILY_TOP_SPOT_BONUS  decimal.Decimal
	WEEKLY_TOP_SPOT_BONUS decimal.Decimal

	POINT_BONUS_SOCIAL_SUBS_TARGET_ACHIEVED decimal.Decimal

	POINT_BONUS_MONTHLY_DAILY_TIME  int64
	POINT_BONUS_MONTHLY_TIME_TARGET int64
	POINT_BONUS_MONTHLY_TIME_REWARD decimal.Decimal
	POINT_BONUS_MONTHLY_TIME_SKIP   int64
	BONUS_MONTHLY_TIME_ENABLE       bool

	TOKENOMICS_V2_PAID_VIEW_TIME_SECONDS        int
	TOKENOMICS_V2_PAID_SESSION_TTL_SECONDS      int
	BONUS_MONTHLY_TIME_DO_NOT_MISS_TIME_MINUTES int

	FIRST_SOCIAL_MEDIA_QUANTITY              int
	FIRST_N_SOCIAL_MEDIA_BONUS_POINTS_AMOUNT decimal.Decimal
	CURRENT_TOKENOMICS_VERSION               int

	WITHDRAWAL_USER_WEEKLY_LIMIT decimal.Decimal
	SURVEY_SECRET_HASH           string
	SURVEY_POINTS                decimal.Decimal

	TOTAL_HOURS_REFERRED_WATCH_VIDEO        int
	POINTS_BONUS_REFERRED_WATCH_VIDEO       decimal.Decimal
	MAX_POINTS_BONUS_REFERRED_WATCH_VIDEO   decimal.Decimal
	TOTAL_FRIENDS_REFERRED_IN_7_DAYS        int
	POINTS_BONUS_FRIENDS_REFERRED_IN_7_DAYS decimal.Decimal

	MONTHLY_CAPPED_POINT_WELCOME_TIER                      decimal.Decimal
	MONTHLY_WITHDRAWAL_LIMIT_WELCOME_TIER                  int
	MONTHLY_POINT_REQUIREMENT_WELCOME_TIER                 decimal.Decimal
	MONTHLY_AD_WATCH_REQUIREMENT_WELCOME_TIER              int64
	WEEKLY_TIME_SPENT_REQUIREMENT_WELCOME_TIER             int
	WEEKLY_TIME_SPENT_PERCENT_POINT_DEDUCTION_WELCOME_TIER int

	MONTHLY_PERCENT_POINT_BEGINNERS_TIER                     int
	MONTHLY_WITHDRAWAL_LIMIT_BEGINNERS_TIER                  int
	MONTHLY_POINT_REQUIREMENT_BEGINNERS_TIER                 decimal.Decimal
	MONTHLY_AD_WATCH_REQUIREMENT_BEGINNERS_TIER              int64
	WEEKLY_TIME_SPENT_REQUIREMENT_BEGINNERS_TIER             int
	WEEKLY_TIME_SPENT_PERCENT_POINT_DEDUCTION_BEGINNERS_TIER int

	MONTHLY_PERCENT_POINT_INTERMEDIARY_TIER                     int
	MONTHLY_WITHDRAWAL_LIMIT_INTERMEDIARY_TIER                  int
	MONTHLY_POINT_REQUIREMENT_INTERMEDIARY_TIER                 decimal.Decimal
	MONTHLY_AD_WATCH_REQUIREMENT_INTERMEDIARY_TIER              int64
	WEEKLY_TIME_SPENT_REQUIREMENT_INTERMEDIARY_TIER             int
	WEEKLY_TIME_SPENT_PERCENT_POINT_DEDUCTION_INTERMEDIARY_TIER int

	MONTHLY_PERCENT_POINT_ADVANCED_TIER                     int
	MONTHLY_WITHDRAWAL_LIMIT_ADVANCED_TIER                  int
	MONTHLY_POINT_REQUIREMENT_ADVANCED_TIER                 decimal.Decimal
	MONTHLY_AD_WATCH_REQUIREMENT_ADVANCED_TIER              int64
	WEEKLY_TIME_SPENT_REQUIREMENT_ADVANCED_TIER             int
	WEEKLY_TIME_SPENT_PERCENT_POINT_DEDUCTION_ADVANCED_TIER int

	MONTHLY_PERCENT_POINT_SUPERSTAR_TIER                     int
	MONTHLY_WITHDRAWAL_LIMIT_SUPERSTAR_TIER                  int
	MONTHLY_POINT_REQUIREMENT_SUPERSTAR_TIER                 decimal.Decimal
	MONTHLY_AD_WATCH_REQUIREMENT_SUPERSTAR_TIER              int64
	WEEKLY_TIME_SPENT_REQUIREMENT_SUPERSTAR_TIER             int
	WEEKLY_TIME_SPENT_PERCENT_POINT_DEDUCTION_SUPERSTAR_TIER int

	COUNTRY_RATE_CONVERSION_ENABLED bool
	WITHDRAWAL_RATE                 decimal.Decimal
}
