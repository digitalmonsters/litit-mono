package solana_api_gate

type TransferTokenResponseData struct {
	TransferredAmount string `json:"transferred_amount"`
	Sender            string `json:"sender"`
	ATA               string `json:"ata"`
	Recipient         string `json:"recipient"`
	FundingSpent      string `json:"funding_spent"`
	Signature         string `json:"signature"`
}

type TransferRequest struct {
	WithdrawalTransactionId int64      `json:"withdrawal_transaction_id"`
	From                    string     `json:"from"`
	Amount                  string     `json:"amount"`
	To                      *Recipient `json:"to"`
}

type Recipient struct {
	Account string `json:"account"`
	Type    string `json:"type"`
}

type CreateVestingRequest struct {
	WithdrawalTransactionId int64  `json:"withdrawal_transaction_id"`
	From                    string `json:"from"`
	To                      string `json:"to"`
	Amounts                 string `json:"amounts"`
	Timestamps              string `json:"timestamps"`
}

type CreateVestingResponseData struct {
	Seed                 string `json:"seed"`
	VestingAccountPubkey string `json:"vesting_account_pubkey"`
}
