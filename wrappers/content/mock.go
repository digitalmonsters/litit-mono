package content

import "go.elastic.co/apm"

type ContentWrapperMock struct {
	GetInternalFn             func(contentIds []int64, includeDeleted bool, apmTransaction *apm.Transaction, forceLog bool) chan ContentGetInternalResponseChan
	GetTopNotFollowingUsersFn func(userId int64, limit int, offset int, apmTransaction *apm.Transaction, forceLog bool) chan GetTopNotFollowingUsersResponseChan
}

func (w *ContentWrapperMock) GetInternal(contentIds []int64, includeDeleted bool, apmTransaction *apm.Transaction, forceLog bool) chan ContentGetInternalResponseChan {
	return w.GetInternalFn(contentIds, includeDeleted, apmTransaction, forceLog)
}

func (w *ContentWrapperMock) GetTopNotFollowingUsers(userId int64, limit int, offset int, apmTransaction *apm.Transaction, forceLog bool) chan GetTopNotFollowingUsersResponseChan {
	return w.GetTopNotFollowingUsersFn(userId, limit, offset, apmTransaction, forceLog)
}

func GetMock() IContentWrapper { // for compiler errors
	return &ContentWrapperMock{}
}
