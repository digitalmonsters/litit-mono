package watch

import "go.elastic.co/apm"

//goland:noinspection ALL
type WatchWrapperMock struct {
	GetLastWatchesByUserFn func(userIds []int64, limitPerUser int, apmTransaction *apm.Transaction, forceLog bool) chan LastWatcherByUserResponseChan
}

func (m *WatchWrapperMock) GetLastWatchesByUsers(userIds []int64, limitPerUser int, apmTransaction *apm.Transaction, forceLog bool) chan LastWatcherByUserResponseChan {
	return m.GetLastWatchesByUserFn(userIds, limitPerUser, apmTransaction, forceLog)
}

func GetMock() IWatchWrapper { // for compiler errors
	return &WatchWrapperMock{}
}
