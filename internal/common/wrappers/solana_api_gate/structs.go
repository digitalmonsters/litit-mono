// [DEPRECATED] not use anymore
package solana_api_gate

import "fmt"

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
	UserId                  int64      `json:"user_id"`
	AdminId                 int64      `json:"admin_id"`
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
	UserId                  int64  `json:"user_id"`
	AdminId                 int64  `json:"admin_id"`
}

type CreateVestingResponseData struct {
	Seed                 string `json:"seed"`
	VestingAccountPubkey string `json:"vesting_account_pubkey"`
}

type TransactionResponseData struct {
	TransactionId int64 `json:"transaction_id"`
}

type TransactionDetail struct {
	Id           int64             `json:"id"`
	BlockchainTx string            `json:"blockchain_tx"`
	Status       TransactionStatus `json:"status"`
}

type GetTransactionsStatusRequest struct {
	WithdrawalIds []int64 `json:"withdrawal_ids"`
}

type TransactionStatus int

const (
	TransactionStatusPending           TransactionStatus = 1
	TransactionStatusSentForProcessing TransactionStatus = 2
	TransactionStatusPaid              TransactionStatus = 3
	TransactionStatusFailed            TransactionStatus = 4
	TransactionStatusTechnicalFail     TransactionStatus = 5
)

func (s TransactionStatus) ToString() string {
	switch s {
	case TransactionStatusPending:
		return "pending"
	case TransactionStatusSentForProcessing:
		return "sent for processing"
	case TransactionStatusPaid:
		return "paid"
	case TransactionStatusFailed:
		return "failed"
	case TransactionStatusTechnicalFail:
		return "technical fail"
	default:
		return fmt.Sprint(s)
	}
}

type TransactionType int

const (
	TransactionTypeTransfer TransactionType = 1
	TransactionTypeVesting  TransactionType = 2
)

func (s TransactionType) ToString() string {
	switch s {
	case TransactionTypeTransfer:
		return "transfer"
	case TransactionTypeVesting:
		return "vesting"
	default:
		return fmt.Sprint(s)
	}
}
