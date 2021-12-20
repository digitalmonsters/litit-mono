package follow

import "go.elastic.co/apm"

//goland:noinspection GoNameStartsWithPackageName
type FollowWrapperMock struct {
	GetUserFollowingRelationBulkFn func(userId int64, requestUserIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUserFollowingRelationBulkResponseChan
	GetUserFollowingRelationFn     func(userId int64, requestUserId int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUserFollowingRelationResponseChan
	GetUserFollowersFn             func(userId int64, pageState string, limit int, apmTransaction *apm.Transaction, forceLog bool) chan GetUserFollowersResponseChan
	GetFollowersCountFn            func(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetFollowersCountResponseChan
}

func (w *FollowWrapperMock) GetUserFollowingRelationBulk(userId int64, requestUserIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUserFollowingRelationBulkResponseChan {
	return w.GetUserFollowingRelationBulkFn(userId, requestUserIds, apmTransaction, forceLog)
}

func (w *FollowWrapperMock) GetUserFollowingRelation(userId int64, requestUserId int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUserFollowingRelationResponseChan {
	return w.GetUserFollowingRelationFn(userId, requestUserId, apmTransaction, forceLog)
}

func (w *FollowWrapperMock) GetUserFollowers(userId int64, pageState string, limit int, apmTransaction *apm.Transaction, forceLog bool) chan GetUserFollowersResponseChan {
	return w.GetUserFollowersFn(userId, pageState, limit, apmTransaction, forceLog)
}

func (w *FollowWrapperMock) GetFollowersCount(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetFollowersCountResponseChan {
	return w.GetFollowersCountFn(userIds, apmTransaction, forceLog)
}

func GetMock() IFollowWrapper { // for compiler errors
	return &FollowWrapperMock{}
}
