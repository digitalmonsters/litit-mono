package user_go

import "go.elastic.co/apm"

//goland:noinspection ALL
type UserGoWrapperMock struct {
	GetUsersFn                   func(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersResponseChan
	GetUsersDetailFn             func(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersDetailsResponseChan
	GetProfileBulkFn             func(currentUserId int64, userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetProfileBulkResponseChan
	GetUsersActiveThresholdsFn   func(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersActiveThresholdsResponseChan
	GetUserIdsFilterByUsernameFn func(userIds []int64, searchQuery string, apmTransaction *apm.Transaction, forceLog bool) chan GetUserIdsFilterByUsernameResponseChan
	GetUsersTagsFn               func(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersTagsResponseChan
	AuthGuestFn                  func(deviceId string, apmTransaction *apm.Transaction, forceLog bool) chan AuthGuestResponseChan
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

func (m *UserGoWrapperMock) GetUsersActiveThresholds(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersActiveThresholdsResponseChan {
	return m.GetUsersActiveThresholdsFn(userIds, apmTransaction, forceLog)
}

func (m *UserGoWrapperMock) GetUserIdsFilterByUsername(userIds []int64, searchQuery string, apmTransaction *apm.Transaction, forceLog bool) chan GetUserIdsFilterByUsernameResponseChan {
	return m.GetUserIdsFilterByUsernameFn(userIds, searchQuery, apmTransaction, forceLog)
}

func (m *UserGoWrapperMock) GetUsersTags(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersTagsResponseChan {
	return m.GetUsersTagsFn(userIds, apmTransaction, forceLog)
}

func (m *UserGoWrapperMock) AuthGuest(deviceId string, apmTransaction *apm.Transaction, forceLog bool) chan AuthGuestResponseChan {
	return m.AuthGuestFn(deviceId, apmTransaction, forceLog)
}

func GetMock() IUserGoWrapper { // for compiler errors
	return &UserGoWrapperMock{}
}
