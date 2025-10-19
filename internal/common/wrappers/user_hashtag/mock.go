package user_hashtag

import "go.elastic.co/apm"

//goland:noinspection ALL
type UserHashtagWrapperMock struct {
	GetUserHashtagSubscriptionStateBulkFn func(hashtags []string, userId int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUserHashtagSubscriptionStateResponseChan
}

func (m *UserHashtagWrapperMock) GetUserHashtagSubscriptionStateBulk(hashtags []string, userId int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUserHashtagSubscriptionStateResponseChan {
	return m.GetUserHashtagSubscriptionStateBulkFn(hashtags, userId, apmTransaction, forceLog)
}

func GetMock() IUserHashtagWrapper { // for compiler errors
	return &UserHashtagWrapperMock{}
}
