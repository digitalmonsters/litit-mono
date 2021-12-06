package user_block

import "go.elastic.co/apm"

//goland:noinspection ALL
type UserBlockWrapperMock struct {
	GetUserBlockFn func(blockedTo int64, blockedBy int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUserBlockResponseChan
}

func (m *UserBlockWrapperMock) GetUserBlock(blockedTo int64, blockedBy int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUserBlockResponseChan {
	return m.GetUserBlockFn(blockedTo, blockedBy, apmTransaction, forceLog)
}

func GetMock() IUserBlockWrapper { // for compiler errors
	return &UserBlockWrapperMock{}
}
