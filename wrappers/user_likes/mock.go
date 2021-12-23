package user_likes

import "go.elastic.co/apm"

//goland:noinspection ALL
type UserLikesWrapperMock struct {
	GetUserLikesFn func(userId int64, limit int, offset int, apmTransaction *apm.Transaction, forceLog bool) chan GetUserLikesResponseChan
}

func (m *UserLikesWrapperMock) GetUserLikes(userId int64, limit int, offset int, apmTransaction *apm.Transaction, forceLog bool) chan GetUserLikesResponseChan {
	return m.GetUserLikesFn(userId, limit, offset, apmTransaction, forceLog)
}

func GetMock() IUserLikesWrapper { // for compiler errors
	return &UserLikesWrapperMock{}
}
