package go_tokenomics

import (
	"fmt"
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
	PointsEarnedTypeNone                     = PointsEarnedType(0)
	PointsEarnedTypeBonusVerify              = PointsEarnedType(1)  // "extra bonus"
	PointsEarnedTypeBonusVerifyReferrer      = PointsEarnedType(2)  // "bonus verify referer"
	PointsEarnedTypeTipSent                  = PointsEarnedType(3)  // "tip sent"
	PointsEarnedTypeTipReceived              = PointsEarnedType(4)  // "tip received"
	PointsEarnedTypeWithdraw                 = PointsEarnedType(5)  // "withdraw"
	PointsEarnedTypeDailyTime                = PointsEarnedType(6)  // "daily time"
	PointsEarnedTypeUserView                 = PointsEarnedType(7)  // "user view"
	PointsEarnedTypeSharerView               = PointsEarnedType(8)  // "sharer view"
	PointsEarnedTypeCreatorView              = PointsEarnedType(9)  // "creator view"
	PointsEarnedTypeBonusPerformance         = PointsEarnedType(10) // "bonus performance"
	PointsEarnedTypeDailyFollowers           = PointsEarnedType(11) // "daily followers"
	PointsEarnedTypeUserShare                = PointsEarnedType(12) // "user share"
	PointsEarnedTypeVerifyGrandReferer       = PointsEarnedType(13) // "bonus verify grand referer"
	PointsEarnedTypeAdmin                    = PointsEarnedType(14) // "admin"
	PointsEarnedTypeCreatorShare             = PointsEarnedType(15) // "creator share"
	PointsEarnedTypeCreatorShareGuest        = PointsEarnedType(16) // "creator share guest"
	PointsEarnedTypeKycPayment               = PointsEarnedType(17) // "kyc payment"
	PointsEarnedTypeBonusKyc                 = PointsEarnedType(18) // "bonus kyc"
	PointsEarnedTypeCreatorViewGuest         = PointsEarnedType(19) // "creator view guest"
	PointsEarnedTypeTapJoy                   = PointsEarnedType(20) // "tapjoy"
	PointsEarnedTypeWeeklyTime               = PointsEarnedType(21) // "weekly time"
	PointsEarnedTypeWeeklyFollowers          = PointsEarnedType(22) // "weekly followers"
	PointsEarnedTypeSharerViewGuest          = PointsEarnedType(23) // "sharer view guest"
	PointsEarnedTypeTransfer                 = PointsEarnedType(24) // new type
	PointsEarnedTypeFeatureBought            = PointsEarnedType(25)
	PointsEarnedTypeApplovin                 = PointsEarnedType(26) // "applovin"
	PointsEarnedTypeWithdrawRejected         = PointsEarnedType(27)
	PointsEarnedTypeMegaBonus                = PointsEarnedType(28) // "mega bonus"
	PointsEarnedTypeAvatarAdded              = PointsEarnedType(29) // "avatar added"
	PointsEarnedTypeDescription              = PointsEarnedType(30) // "adding description"
	PointsEarnedTypeUploadFirstVideo         = PointsEarnedType(31)
	PointsEarnedTypePointsWriteOff           = PointsEarnedType(32) // write off money
	PointsEarnedTypeTechnicTransfer          = PointsEarnedType(33) //technical transfer
	PointsEarnedTypeUploadFirstSpot          = PointsEarnedType(34)
	PointsEarnedTypeSpotsUserView            = PointsEarnedType(35)
	PointsEarnedTypeSpotsCreatorView         = PointsEarnedType(36)
	PointsEarnedTypeSpotsCreatorViewGuest    = PointsEarnedType(37)
	PointsEarnedTypeSpotsBonusPerformance    = PointsEarnedType(38)
	PointsEarnedTypeEmailMarketingAdded      = PointsEarnedType(39) // "email marketing added"
	PointsEarnedTypeIronsource               = PointsEarnedType(40) // "ironsource"
	PointsEarnedTypeTopSpotDaily             = PointsEarnedType(41)
	PointsEarnedTypeTopSpotWeekly            = PointsEarnedType(42)
	PointsEarnedTypeSocialSubsTargetAchieved = PointsEarnedType(43)
)

type WithdrawalStatus int16

const (
	WithdrawalStatusNone                         WithdrawalStatus = 0
	WithdrawalStatusPending                      WithdrawalStatus = 1
	WithdrawalStatusApproved                     WithdrawalStatus = 2
	WithdrawalStatusRejected                     WithdrawalStatus = 3 // final
	WithdrawalStatusFailed                       WithdrawalStatus = 4 // final
	WithdrawalStatusPaymentPending               WithdrawalStatus = 5
	WithdrawalStatusPaid                         WithdrawalStatus = 6 // final
	WithdrawalStatusPaymentInvestigationRequired WithdrawalStatus = 7
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
type GetReferralInfoResponse struct {
	TotalCollectedPoints      decimal.Decimal           `json:"total_collected_points"`
	Referrals                 map[int64]decimal.Decimal `json:"referrals"`
	TotalGrandCollectedPoints decimal.Decimal           `json:"total_grand_collected_points"`
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
	FirstSpotUploaded        bool `json:"first_spot_uploaded"`
	FirstTimeAvatarAdded     bool `json:"first_time_avatar_added"`
	FirstEmailMarketingAdded bool `json:"first_email_marketing_added"`
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
