package like

import "go.elastic.co/apm"

//goland:noinspection ALL
type LikeWrapperMock struct {
	GetLastLikesByUsersFn func(userIds []int64, limitPerUser int, apmTransaction *apm.Transaction, forceLog bool) chan LastLikedByUserResponseChan
	GetLikeContentUserByContentIdsInternalFn func(contentIds []int64, userId int64, apmTransaction *apm.Transaction, forceLog bool) chan LikeContentUserByContentIdsResponseChan
}

func (m *LikeWrapperMock) GetLastLikesByUsers(userIds []int64, limitPerUser int, apmTransaction *apm.Transaction, forceLog bool) chan LastLikedByUserResponseChan {
	return m.GetLastLikesByUsersFn(userIds, limitPerUser, apmTransaction, forceLog)
}

func (w *LikeWrapperMock) GetLikeContentUserByContentIdsInternal(contentIds []int64, userId int64, apmTransaction *apm.Transaction, forceLog bool) chan LikeContentUserByContentIdsResponseChan {
	return w.GetLikeContentUserByContentIdsInternalFn(contentIds, userId, apmTransaction, forceLog)
}


func GetMock() ILikeWrapper { // for compiler errors
	return &LikeWrapperMock{}
}
