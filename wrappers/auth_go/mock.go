package auth_go

import (
	"context"

	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/wrappers"
	"go.elastic.co/apm"
)

type AuthGoWrapperMock struct {
	CheckAdminPermissionsFn         func(userId int64, obj string, transaction *apm.Transaction, forceLog bool) chan CheckAdminPermissionsResponseChan
	CheckLegacyAdminFn              func(userId int64, transaction *apm.Transaction, forceLog bool) chan CheckLegacyAdminResponseChan
	GetAdminIdsFilterByEmailFn      func(adminIds []int64, searchQuery string, apmTransaction *apm.Transaction, forceLog bool) chan GetAdminIdsFilterByEmailResponseChan
	GetAdminsInfoByIdFn             func(adminIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetAdminsInfoByIdResponseChan
	AddNewUserFn                    func(req eventsourcing.UserEvent, apmTransaction *apm.Transaction, forceLog bool) chan AddUserResponseChan
	IsGuestFn                       func(userId int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[IsGuestResponse]
	GetUsersRegistrationTypeFn      func(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[int64]SocialProviderType]
	InternalGetUsersForValidationFn func(userIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]UserForValidator]
	UpdateEmailForUserFn            func(userId int64, email string, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]UpdateEmailForUserResponse]
	GetOnlineUsersFn                func(forceLog bool) chan wrappers.GenericResponseChan[OnlineUserResponse]
}

// GetOnlineUsers implements IAuthGoWrapper.
func (w *AuthGoWrapperMock) GetOnlineUsers(forceLog bool) chan wrappers.GenericResponseChan[OnlineUserResponse] {
	return w.GetOnlineUsersFn(forceLog)
}

// TriggerUserOnline implements IAuthGoWrapper.
func (w *AuthGoWrapperMock) TriggerUserOnline(userId int64) chan wrappers.GenericResponseChan[GenericTriggerOnlineOfflineRequest] {
	panic("unimplemented")
}

func (w *AuthGoWrapperMock) TriggerUserOffline(userId int64) chan wrappers.GenericResponseChan[GenericTriggerOnlineOfflineRequest] {
	panic("unimplemented")
}

func (w *AuthGoWrapperMock) InternalGetUsersForValidation(userIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]UserForValidator] {
	return w.InternalGetUsersForValidationFn(userIds, ctx, forceLog)
}

func (w *AuthGoWrapperMock) IsGuest(userId int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[IsGuestResponse] {
	return w.IsGuestFn(userId, apmTransaction, forceLog)
}

func (w *AuthGoWrapperMock) UpdateEmailForUser(userId int64, email string, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]UpdateEmailForUserResponse] {
	return w.UpdateEmailForUserFn(userId, email, ctx, forceLog)
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
