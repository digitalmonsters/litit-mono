package follow

import "go.elastic.co/apm"

//goland:noinspection GoNameStartsWithPackageName
type FollowWrapperMock struct {
	GetFollowContentUserByContentIdsInternalFn func(contentIds []int64, userId int64, apmTransaction *apm.Transaction, forceLog bool) chan FollowContentUserByContentIdsResponseChan
}

func (w *FollowWrapperMock) GetFollowContentUserByContentIdsInternal(contentIds []int64, userId int64, apmTransaction *apm.Transaction, forceLog bool) chan FollowContentUserByContentIdsResponseChan {
	return w.GetFollowContentUserByContentIdsInternalFn(contentIds, userId, apmTransaction, forceLog)
}

func GetMock() IFollowWrapper { // for compiler errors
	return &FollowWrapperMock{}
}
