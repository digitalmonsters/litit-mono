package like

import "go.elastic.co/apm"

//goland:noinspection ALL
type LikeWrapperMock struct {
	GetLastLikesByUsersFn func(userIds []int64, limitPerUser int, apmTransaction *apm.Transaction, forceLog bool) chan LastLikedByUserResponse
}

func (m *LikeWrapperMock) GetLastLikesByUsers(userIds []int64, limitPerUser int, apmTransaction *apm.Transaction, forceLog bool) chan LastLikedByUserResponse {
	return m.GetLastLikesByUsersFn(userIds, limitPerUser, apmTransaction, forceLog)
}

func GetMock() ILikeWrapper { // for compiler errors
	return &LikeWrapperMock{}
}
