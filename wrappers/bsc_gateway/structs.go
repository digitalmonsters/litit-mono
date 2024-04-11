package bsc_gateway

type SignatureResponseData struct {
	From                    string `json:"from"`
	Amount                  string `json:"amount"`
	Signature               string `json:"signature"`
	CommunityPoolSMC        string `json:"community_pool_smc"`
	WithdrawalTransactionId int64  `json:"withdrawal_transaction_id"`
	UserId                  int64  `json:"user_id"`
	AdminId                 int64  `json:"admin_id"`
	ExpiredAt               int64  `json:"expired_at"`
}

type CreateSignatureRequest struct {
	From                    string `json:"from"`
	Amount                  string `json:"amount"`
	WithdrawalTransactionId int64  `json:"withdrawal_transaction_id"`
	UserId                  int64  `json:"user_id"`
	AdminId                 int64  `json:"admin_id"`
}

type GetSignaturesRequest struct {
	WithdrawalIds []int64 `json:"withdrawal_ids"`
}
