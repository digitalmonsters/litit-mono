package user

import "go.elastic.co/apm"

//goland:noinspection ALL
type UserWrapperMock struct {
	GetUsersFn       func(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersResponseChan
	GetUsersDetailFn func(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersDetailsResponseChan
}

func (m *UserWrapperMock) GetUsers(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersResponseChan {
	return m.GetUsersFn(userIds, apmTransaction, forceLog)
}

func (m *UserWrapperMock) GetUsersDetails(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersDetailsResponseChan {
	return m.GetUsersDetailFn(userIds, apmTransaction, forceLog)
}

func GetMock() IUserWrapper { // for compiler errors
	return &UserWrapperMock{}
}
