package follow

import "go.elastic.co/apm"

//goland:noinspection GoNameStartsWithPackageName
type FollowWrapperMock struct {
	GetFollowContentUserByContentAuthorIdsInternalFn func(contentAuthorIds []int64, userId int64, apmTransaction *apm.Transaction, forceLog bool) chan FollowContentUserByContentAuthorIdsResponseChan
}

func (w *FollowWrapperMock) GetFollowContentUserByContentAuthorIdsInternal(contentAuthorIds []int64, userId int64, apmTransaction *apm.Transaction, forceLog bool) chan FollowContentUserByContentAuthorIdsResponseChan {
	return w.GetFollowContentUserByContentAuthorIdsInternalFn(contentAuthorIds, userId, apmTransaction, forceLog)
}

func GetMock() IFollowWrapper { // for compiler errors
	return &FollowWrapperMock{}
}
