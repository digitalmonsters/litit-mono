package auth_go

import (
	"go.elastic.co/apm"
)

type AuthGoWrapperMock struct {
	CheckAdminPermissionsFn func(userId int64, obj string, transaction *apm.Transaction, forceLog bool) chan CheckAdminPermissionsResponseChan
	CheckLegacyAdminFn      func(userId int64, transaction *apm.Transaction, forceLog bool) chan CheckLegacyAdminResponseChan
}

func (w *AuthGoWrapperMock) CheckLegacyAdmin(userId int64, transaction *apm.Transaction, forceLog bool) chan CheckLegacyAdminResponseChan {
	return w.CheckLegacyAdminFn(userId, transaction, forceLog)
}

func (w *AuthGoWrapperMock) CheckAdminPermissions(userId int64, obj string, transaction *apm.Transaction, forceLog bool) chan CheckAdminPermissionsResponseChan {
	return w.CheckAdminPermissionsFn(userId, obj, transaction, forceLog)
}

func GetMock() IAuthGoWrapper { // for compiler errors
	return &AuthGoWrapperMock{}
}
