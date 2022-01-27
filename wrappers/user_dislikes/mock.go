package user_dislikes

import "go.elastic.co/apm"

//goland:noinspection ALL
type UserDislikesWrapperMock struct {
	GetAllUserDislikesFn func(userId int64, apmTransaction *apm.Transaction, forceLog bool) chan GetAllUserDislikesResponseChan
}

func (m *UserDislikesWrapperMock) GetAllUserDislikes(userId int64, apmTransaction *apm.Transaction, forceLog bool) chan GetAllUserDislikesResponseChan {
	return m.GetAllUserDislikesFn(userId, apmTransaction, forceLog)
}

func GetMock() IUserDislikesWrapper { // for compiler errors
	return &UserDislikesWrapperMock{}
}
