package user_go

import "go.elastic.co/apm"

//goland:noinspection ALL
type UserGoWrapperMock struct {
	GetUsersFn       func(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersResponseChan
	GetUsersDetailFn func(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersDetailsResponseChan
	GetProfileBulkFn func(currentUserId int64, userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetProfileBulkResponseChan
}

func (m *UserGoWrapperMock) GetUsers(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersResponseChan {
	return m.GetUsersFn(userIds, apmTransaction, forceLog)
}

func (m *UserGoWrapperMock) GetUsersDetails(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersDetailsResponseChan {
	return m.GetUsersDetailFn(userIds, apmTransaction, forceLog)
}

func (m *UserGoWrapperMock) GetProfileBulk(currentUserId int64, userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetProfileBulkResponseChan {
	return m.GetProfileBulkFn(currentUserId, userIds, apmTransaction, forceLog)
}

func GetMock() IUserGoWrapper { // for compiler errors
	return &UserGoWrapperMock{}
}

