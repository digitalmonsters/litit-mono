package go_tokenomics

import (
	"github.com/digitalmonsters/go-common/filters"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/shopspring/decimal"
)

type GetUsersTokenomicsInfoRequest struct {
	UserIds []int64          `json:"user_ids"`
	Filters []filters.Filter `json:"filters"`
}

type GetUsersTokenomicsInfoResponseChan struct {
	Items map[int64]UserTokenomicsInfo
	Error *rpc.RpcError
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

type UserTokenomicsStats struct {
	LITITBalance                  decimal.Decimal `json:"litit_balance"`
	PointsForViews                decimal.Decimal `json:"points_for_views"`
	TipsNumber                    int             `json:"tips_number"`
	PointsForTips                 decimal.Decimal `json:"points_for_tips"`
	TapjoyActivityNumber          int             `json:"tapjoy_activity_number"`
	PointsForTapjoyActivity       decimal.Decimal `json:"points_for_tapjoy_activity"`
	PointsForInviting             decimal.Decimal `json:"points_for_inviting"`
	ApprovedTransactionsNumber    int             `json:"approved_transactions_number"`
	PointsForApprovedTransactions decimal.Decimal `json:"points_for_approved_transactions"`
	RejectedTransactionsNumber    int             `json:"rejected_transactions_number"`
	PointsForRejectedTransactions decimal.Decimal `json:"points_for_rejected_transactions"`
	PendingTransactionsNumber     int             `json:"pending_transactions_number"`
	PointsForPendingTransactions  decimal.Decimal `json:"points_for_pending_transactions"`
	SharedVideoNumber             int             `json:"shared_video_number"`
	PointsForSharedVideo          decimal.Decimal `json:"points_for_shared_video"`
	InvitedFromShareNumber        int             `json:"invited_from_share_number"`
	PointsForInvitedFromShare     decimal.Decimal `json:"points_for_invited_from_share"`
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
	TotalCollectedPoints decimal.Decimal           `json:"total_collected_points"`
	Referrals            map[int64]decimal.Decimal `json:"referrals"`
}

type GetActivitiesInfoRequest struct {
	UserId int64 `json:"user_id"`
}

type GetActivitiesInfoResponse struct {
	Items map[int64]UserActivity `json:"items"`
	Error *rpc.RpcError
}

type UserActivity struct {
	AddDescriptionBonus  bool `json:"add_description_bonus"`
	FirstVideoUploaded   bool `json:"first_video_uploaded"`
	FirstTimeAvatarAdded bool `json:"first_time_avatar_added"`
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
