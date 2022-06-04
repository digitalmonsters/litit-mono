package like

import (
	"context"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/wrappers"
	"go.elastic.co/apm"
)

//goland:noinspection ALL
type LikeWrapperMock struct {
	GetInternalLikedByUserFn         func(contentIds []int64, userId int64, apmTransaction *apm.Transaction, forceLog bool) chan GetInternalLikedByUserResponseChan
	GetInternalDislikedByUserFn      func(contentIds []int64, userId int64, apmTransaction *apm.Transaction, forceLog bool) chan GetInternalDislikedByUserResponseChan
	GetInternalSpotReactionsByUserFn func(contentIds []int64, userId int64, apmTransaction *apm.Transaction, forceLog bool) chan GetInternalSpotReactionsByUserResponseChan

	GetLastLikesByUsersFn  func(userIds []int64, limitPerUser int, apmTransaction *apm.Transaction, forceLog bool) chan LastLikedByUserResponseChan
	GetInternalUserLikesFn func(userId int64, size int, pageState string, apmTransaction *apm.Transaction, forceLog bool) chan GetInternalUserLikesResponseChan
	AddLikesInternalFn     func(likeEvents []eventsourcing.LikeEvent, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[AddLikesResponse]
}

func (w *LikeWrapperMock) GetInternalLikedByUser(contentIds []int64, userId int64, apmTransaction *apm.Transaction, forceLog bool) chan GetInternalLikedByUserResponseChan {
	return w.GetInternalLikedByUserFn(contentIds, userId, apmTransaction, forceLog)
}
func (w *LikeWrapperMock) GetInternalDislikedByUser(contentIds []int64, userId int64, apmTransaction *apm.Transaction, forceLog bool) chan GetInternalDislikedByUserResponseChan {
	return w.GetInternalDislikedByUserFn(contentIds, userId, apmTransaction, forceLog)
}
func (w *LikeWrapperMock) GetInternalSpotReactionsByUser(contentIds []int64, userId int64, apmTransaction *apm.Transaction, forceLog bool) chan GetInternalSpotReactionsByUserResponseChan {
	return w.GetInternalSpotReactionsByUserFn(contentIds, userId, apmTransaction, forceLog)
}

func (w *LikeWrapperMock) GetLastLikesByUsers(userIds []int64, limitPerUser int, apmTransaction *apm.Transaction, forceLog bool) chan LastLikedByUserResponseChan {
	return w.GetLastLikesByUsersFn(userIds, limitPerUser, apmTransaction, forceLog)
}
func (w *LikeWrapperMock) GetInternalUserLikes(userId int64, size int, pageState string, apmTransaction *apm.Transaction, forceLog bool) chan GetInternalUserLikesResponseChan {
	return w.GetInternalUserLikesFn(userId, size, pageState, apmTransaction, forceLog)
}
func (w *LikeWrapperMock) AddLikesInternal(likeEvents []eventsourcing.LikeEvent, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[AddLikesResponse] {
	return w.AddLikesInternalFn(likeEvents, ctx, forceLog)
}

func GetMock() ILikeWrapper { // for compiler errors
	return &LikeWrapperMock{}
}
