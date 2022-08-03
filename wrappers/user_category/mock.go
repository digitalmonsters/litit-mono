package user_category

import (
	"context"
	"github.com/digitalmonsters/go-common/wrappers"
	"go.elastic.co/apm"
)

//goland:noinspection ALL
type UserCategoryWrapperMock struct {
	GetUserCategorySubscriptionStateBulkFn func(categoryIds []int64, userId int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUserCategorySubscriptionStateResponseChan
	GetInternalUserCategorySubscriptionsFn func(userId int64, limit int, pageState string,
		ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[GetInternalUserCategorySubscriptionsResponse]
}

func (m *UserCategoryWrapperMock) GetUserCategorySubscriptionStateBulk(categoryIds []int64, userId int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUserCategorySubscriptionStateResponseChan {
	return m.GetUserCategorySubscriptionStateBulkFn(categoryIds, userId, apmTransaction, forceLog)
}

func (m *UserCategoryWrapperMock) GetInternalUserCategorySubscriptions(userId int64, limit int, pageState string,
	ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[GetInternalUserCategorySubscriptionsResponse] {
	return m.GetInternalUserCategorySubscriptionsFn(userId, limit, pageState, ctx, forceLog)
}

func GetMock() IUserCategoryWrapper { // for compiler errors
	return &UserCategoryWrapperMock{}
}
