package solana_api_gate

import "github.com/digitalmonsters/go-common/rpc"

type TransferTokenResponseChan struct {
	Data  *TransferTokenResponseData `json:"data"`
	Error *rpc.RpcError
}

type TransferTokenResponseData struct {
	TransferredAmount string `json:"transferred_amount"`
	Sender            string `json:"sender"`
	ATA               string `json:"ata"`
	Recipient         string `json:"recipient"`
	FundingSpent      string `json:"funding_spent"`
	Signature         string `json:"signature"`
}

type TransferRequest struct {
	From   string     `json:"from"`
	Amount string     `json:"amount"`
	To     *Recipient `json:"to"`
}

type Recipient struct {
	Account string `json:"account"`
	Type    string `json:"type"`
}

type CreateVestingResponseChan struct {
	Data  *CreateVestingResponseData `json:"data"`
	Error *rpc.RpcError
}

type CreateVestingRequest struct {
	From       string `json:"from"`
	To         string `json:"to"`
	Amounts    string `json:"amounts"`
	Timestamps string `json:"timestamps"`
}

type CreateVestingResponseData struct {
	Seed                 string `json:"seed"`
	VestingAccountPubkey string `json:"vesting_account_pubkey"`
}
