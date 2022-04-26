package watch

import (
	"context"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/wrappers"
	"go.elastic.co/apm"
)

//goland:noinspection ALL
type WatchWrapperMock struct {
	GetLastWatchesByUserFn              func(userIds []int64, limitPerUser int, apmTransaction *apm.Transaction, forceLog bool) chan LastWatcherByUserResponseChan
	AddViewsInternalFn                  func(viewEvents []eventsourcing.ViewEvent, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[AddViewsResponse]
	GetUsersTotalTimeWatchingInternalFn func(userIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]int64]
}

func (m *WatchWrapperMock) GetLastWatchesByUsers(userIds []int64, limitPerUser int, apmTransaction *apm.Transaction, forceLog bool) chan LastWatcherByUserResponseChan {
	return m.GetLastWatchesByUserFn(userIds, limitPerUser, apmTransaction, forceLog)
}

func (m *WatchWrapperMock) AddViewsInternal(viewEvents []eventsourcing.ViewEvent, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[AddViewsResponse] {
	return m.AddViewsInternalFn(viewEvents, ctx, forceLog)
}

func (m *WatchWrapperMock) GetUsersTotalTimeWatchingInternal(userIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]int64] {
	return m.GetUsersTotalTimeWatchingInternalFn(userIds, ctx, forceLog)
}

func GetMock() IWatchWrapper { // for compiler errors
	return &WatchWrapperMock{}
}
