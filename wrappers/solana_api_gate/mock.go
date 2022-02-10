package solana_api_gate

import "go.elastic.co/apm"

//goland:noinspection ALL
type SolanaApiGateWrapperMock struct {
	TransferTokenFn func(from string, amount string, account string, recipientType string, apmTransaction *apm.Transaction, forceLog bool) chan TransferTokenResponseChan
	CreateVestingFn func(from string, to string, amounts string, timestamps string, apmTransaction *apm.Transaction, forceLog bool) chan CreateVestingResponseChan
}

func (m *SolanaApiGateWrapperMock) TransferToken(from string, amount string, account string, recipientType string, apmTransaction *apm.Transaction, forceLog bool) chan TransferTokenResponseChan {
	return m.TransferTokenFn(from, amount, account, recipientType, apmTransaction, forceLog)
}

func (m *SolanaApiGateWrapperMock) CreateVesting(from string, to string, amounts string, timestamps string, apmTransaction *apm.Transaction, forceLog bool) chan CreateVestingResponseChan {
	return m.CreateVestingFn(from, to, amounts, timestamps, apmTransaction, forceLog)
}

func GetMock() ISolanaApiGateWrapper { // for compiler errors
	return &SolanaApiGateWrapperMock{}
}
