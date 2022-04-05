package like

import (
	"github.com/digitalmonsters/go-common/wrappers"
	"go.elastic.co/apm"
)

//goland:noinspection ALL
type LikeWrapperMock struct {
	GetLastLikesByUsersFn       func(userIds []int64, limitPerUser int, apmTransaction *apm.Transaction, forceLog bool) chan LastLikedByUserResponseChan
	GetInternalLikedByUserFn    func(contentIds []int64, userId int64, apmTransaction *apm.Transaction, forceLog bool) chan GetInternalLikedByUserResponseChan
	GetInternalUserLikesFn      func(userId int64, size int, pageState string, apmTransaction *apm.Transaction, forceLog bool) chan GetInternalUserLikesResponseChan
	GetInternalDislikedByUserFn func(contentIds []int64, userId int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[int64]bool]
}

func (w *LikeWrapperMock) GetLastLikesByUsers(userIds []int64, limitPerUser int, apmTransaction *apm.Transaction, forceLog bool) chan LastLikedByUserResponseChan {
	return w.GetLastLikesByUsersFn(userIds, limitPerUser, apmTransaction, forceLog)
}

func (w *LikeWrapperMock) GetInternalLikedByUser(contentIds []int64, userId int64, apmTransaction *apm.Transaction, forceLog bool) chan GetInternalLikedByUserResponseChan {
	return w.GetInternalLikedByUserFn(contentIds, userId, apmTransaction, forceLog)
}
func (w *LikeWrapperMock) GetInternalUserLikes(userId int64, size int, pageState string, apmTransaction *apm.Transaction, forceLog bool) chan GetInternalUserLikesResponseChan {
	return w.GetInternalUserLikesFn(userId, size, pageState, apmTransaction, forceLog)
}
func (w *LikeWrapperMock) GetInternalDislikedByUser(contentIds []int64, userId int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[int64]bool] {
	return w.GetInternalDislikedByUserFn(contentIds, userId, apmTransaction, forceLog)
}

func GetMock() ILikeWrapper { // for compiler errors
	return &LikeWrapperMock{}
}
