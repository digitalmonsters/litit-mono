package bsc_gateway

type SignatureResponseData struct {
	TransferredAmount string `json:"transferred_amount"`
	Sender            string `json:"sender"`
	ATA               string `json:"ata"`
	Recipient         string `json:"recipient"`
	FundingSpent      string `json:"funding_spent"`
	Signature         string `json:"signature"`
}

type SignatureRequest struct {
	WithdrawalTransactionId int64  `json:"withdrawal_transaction_id"`
	From                    string `json:"from"`
	Amount                  string `json:"amount"`
	UserId                  int64  `json:"user_id"`
	AdminId                 int64  `json:"admin_id"`
}

type GetSignaturesRequest struct {
	WithdrawalIds []int64 `json:"withdrawal_ids"`
}
