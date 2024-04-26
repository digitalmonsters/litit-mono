package bsc_gateway

import (
	"github.com/digitalmonsters/go-common/wrappers/go_tokenomics"
	"github.com/shopspring/decimal"
)

type SignatureResponseData struct {
	From                    string `json:"from"`
	Amount                  string `json:"amount"`
	Signature               string `json:"signature"`
	CommunityPoolSMC        string `json:"community_pool_smc"`
	WithdrawalTransactionId int64  `json:"withdrawal_transaction_id"`
	UserId                  int64  `json:"user_id"`
	AdminId                 int64  `json:"admin_id"`
	ExpiredAt               int64  `json:"expired_at"`
	TnxHash                 string `json:"tnx_hash"`
}

type CreateSignatureRequest struct {
	From                    string          `json:"from"`
	Amount                  decimal.Decimal `json:"amount"`
	WithdrawalTransactionId int64           `json:"withdrawal_transaction_id"`
	UserId                  int64           `json:"user_id"`
	AdminId                 int64           `json:"admin_id"`
}

type GetSignatureStatusRequest struct {
	WithdrawalId int64  `json:"withdrawal_id"`
	Signature    string `json:"string"`
}

type SignatureStatusResponseData struct {
	Status            go_tokenomics.WithdrawalStatus `json:"status"`
	SignatureResponse SignatureResponseData          `json:"signature_response"`
}
