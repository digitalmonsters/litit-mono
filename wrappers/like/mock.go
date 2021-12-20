package like

import "go.elastic.co/apm"

//goland:noinspection ALL
type LikeWrapperMock struct {
	GetLastLikesByUsersFn func(userIds []int64, limitPerUser int, apmTransaction *apm.Transaction, forceLog bool) chan LastLikedByUserResponseChan
	GetInternalLikedByUserFn func(contentIds []int64, userId int64, apmTransaction *apm.Transaction, forceLog bool) chan GetInternalLikedByUserResponseChan
}

func (m *LikeWrapperMock) GetLastLikesByUsers(userIds []int64, limitPerUser int, apmTransaction *apm.Transaction, forceLog bool) chan LastLikedByUserResponseChan {
	return m.GetLastLikesByUsersFn(userIds, limitPerUser, apmTransaction, forceLog)
}

func (w *LikeWrapperMock) GetInternalLikedByUser(contentIds []int64, userId int64, apmTransaction *apm.Transaction, forceLog bool) chan GetInternalLikedByUserResponseChan {
	return w.GetInternalLikedByUserFn(contentIds, userId, apmTransaction, forceLog)
}


func GetMock() ILikeWrapper { // for compiler errors
	return &LikeWrapperMock{}
}
