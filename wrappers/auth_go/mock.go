package auth_go

import (
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/wrappers"
	"go.elastic.co/apm"
)

type AuthGoWrapperMock struct {
	CheckAdminPermissionsFn    func(userId int64, obj string, transaction *apm.Transaction, forceLog bool) chan CheckAdminPermissionsResponseChan
	CheckLegacyAdminFn         func(userId int64, transaction *apm.Transaction, forceLog bool) chan CheckLegacyAdminResponseChan
	GetAdminIdsFilterByEmailFn func(adminIds []int64, searchQuery string, apmTransaction *apm.Transaction, forceLog bool) chan GetAdminIdsFilterByEmailResponseChan
	GetAdminsInfoByIdFn        func(adminIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetAdminsInfoByIdResponseChan
	AddNewUserFn               func(req eventsourcing.UserEvent, apmTransaction *apm.Transaction, forceLog bool) chan AddUserResponseChan
	IsGuestFn                  func(userId int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[IsGuestResponse]
	GetUsersRegistrationTypeFn func(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[int64]SocialProviderType]
}

func (w *AuthGoWrapperMock) IsGuest(userId int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[IsGuestResponse] {
	return w.IsGuestFn(userId, apmTransaction, forceLog)
}

func (w *AuthGoWrapperMock) AddNewUser(req eventsourcing.UserEvent, apmTransaction *apm.Transaction, forceLog bool) chan AddUserResponseChan {
	return w.AddNewUserFn(req, apmTransaction, forceLog)
}

func (w *AuthGoWrapperMock) CheckLegacyAdmin(userId int64, transaction *apm.Transaction, forceLog bool) chan CheckLegacyAdminResponseChan {
	return w.CheckLegacyAdminFn(userId, transaction, forceLog)
}

func (w *AuthGoWrapperMock) CheckAdminPermissions(userId int64, obj string, transaction *apm.Transaction, forceLog bool) chan CheckAdminPermissionsResponseChan {
	return w.CheckAdminPermissionsFn(userId, obj, transaction, forceLog)
}

func (w *AuthGoWrapperMock) GetAdminIdsFilterByEmail(adminIds []int64, searchQuery string, apmTransaction *apm.Transaction, forceLog bool) chan GetAdminIdsFilterByEmailResponseChan {
	return w.GetAdminIdsFilterByEmailFn(adminIds, searchQuery, apmTransaction, forceLog)
}

func (w *AuthGoWrapperMock) GetAdminsInfoById(adminIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetAdminsInfoByIdResponseChan {
	return w.GetAdminsInfoByIdFn(adminIds, apmTransaction, forceLog)
}

func (w *AuthGoWrapperMock) GetUsersRegistrationType(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[int64]SocialProviderType] {
	return w.GetUsersRegistrationTypeFn(userIds, apmTransaction, forceLog)
}

func GetMock() IAuthGoWrapper { // for compiler errors
	return &AuthGoWrapperMock{}
}
