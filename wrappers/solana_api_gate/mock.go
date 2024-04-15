// [DEPRECATED] not use anymore
package solana_api_gate

import (
	"context"

	"github.com/digitalmonsters/go-common/wrappers"
)

//goland:noinspection ALL
type SolanaApiGateWrapperMock struct {
	TransferTokenFn         func(from string, amount string, account string, recipientType string, withdrawalTransactionId int64, userId int64, adminId int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[TransactionResponseData]
	CreateVestingFn         func(from string, to string, amounts string, timestamps string, withdrawalTransactionId int64, userId int64, adminId int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[TransactionResponseData]
	GetTransactionsStatusFn func(withdrawalIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]TransactionDetail]
}

func (m *SolanaApiGateWrapperMock) GetTransactionsStatus(withdrawalIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]TransactionDetail] {
	return m.GetTransactionsStatusFn(withdrawalIds, ctx, forceLog)
}

func (m *SolanaApiGateWrapperMock) TransferToken(from string, amount string, account string, recipientType string, withdrawalTransactionId int64, userId int64, adminId int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[TransactionResponseData] {
	return m.TransferTokenFn(from, amount, account, recipientType, withdrawalTransactionId, userId, adminId, ctx, forceLog)
}

func (m *SolanaApiGateWrapperMock) CreateVesting(from string, to string, amounts string, timestamps string, withdrawalTransactionId int64, userId int64, adminId int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[TransactionResponseData] {
	return m.CreateVestingFn(from, to, amounts, timestamps, withdrawalTransactionId, userId, adminId, ctx, forceLog)
}

func GetMock() ISolanaApiGateWrapper { // for compiler errors
	return &SolanaApiGateWrapperMock{}
}
