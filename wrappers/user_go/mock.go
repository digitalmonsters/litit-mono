package user_go

import (
	"context"

	"github.com/digitalmonsters/go-common/wrappers"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
)

//goland:noinspection ALL
type UserGoWrapperMock struct {
	GetUsersFn       func(userIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]UserRecord]
	GetUsersDetailFn func(userIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]UserDetailRecord]
	GetUserDetailsFn func(userId int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[UserDetailRecord]

	GetPetsFn       func(petIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]PetRecord]
	GetPetsDetailFn func(petIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]PetDetailRecord]
	GetPetDetailsFn func(petId int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[PetDetailRecord]
	GetPetsSearchFn func(keywords string, page, count int, ctx context.Context, forceLog bool) chan SearchPetDetailRecordResponseChan

	GetProfileBulkFn                      func(currentUserId int64, userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetProfileBulkResponseChan
	GetUsersActiveThresholdsFn            func(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersActiveThresholdsResponseChan
	GetUserIdsFilterByUsernameFn          func(userIds []int64, searchQuery string, apmTransaction *apm.Transaction, forceLog bool) chan GetUserIdsFilterByUsernameResponseChan
	GetUsersTagsFn                        func(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersTagsResponseChan
	AuthGuestFn                           func(deviceId string, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[AuthGuestResp]
	GetBlockListFn                        func(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[string][]int64]
	GetUserBlockFn                        func(blockedTo int64, blockedBy int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[UserBlockData]
	UpdateUserMetadataAfterRegistrationFn func(request UpdateUserMetaDataRequest, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[UserRecord]
	ForceResetUserWithNewGuestIdentityFn  func(deviceId string, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[ForceResetUserIdentityWithNewGuestResponse]
	VerifyUserFn                          func(userId int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[UserRecord]
	GetAllActiveBotsFn                    func(ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[GetAllActiveBotsResponse]
	GetConfigPropertiesInternalFn         func(properties []string, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[GetConfigPropertiesResponseChan]
	UpdateEmailMarketingFn                func(userId int64, emailMarketing null.String, emailMarketingVerified bool, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[any]
	GenerateDeeplinkFn                    func(urlPath string, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[GenerateDeeplinkResponse]
	CreateExportFn                        func(name string, exportType ExportType, filters interface{}, exportedBy int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[CreateExportResponse]
	FinalizeExportFn                      func(exportId int64, file null.String, err error, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[FinalizeExportResponse]
	GetGrandReferrerIdsFn                 func(ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[[]int64]
	SetSpotsUploadBannedFn                func(userId int64, banned bool, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[any]
	UpdatePetAlbumFn                      func(petId int64, videoId string, userId int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[any]
}

func (m *UserGoWrapperMock) GetUserDetails(userId int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[UserDetailRecord] {
	return m.GetUserDetailsFn(userId, ctx, forceLog)
}

func (m *UserGoWrapperMock) GetPetDetails(petIds int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[PetDetailRecord] {
	return m.GetPetDetailsFn(petIds, ctx, forceLog)
}

func (m *UserGoWrapperMock) VerifyUser(userId int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[UserRecord] {
	return m.VerifyUserFn(userId, ctx, forceLog)
}

func (m *UserGoWrapperMock) ForceResetUserWithNewGuestIdentity(deviceId string, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[ForceResetUserIdentityWithNewGuestResponse] {
	return m.ForceResetUserWithNewGuestIdentityFn(deviceId, ctx, forceLog)
}

func (m *UserGoWrapperMock) UpdateUserMetadataAfterRegistration(request UpdateUserMetaDataRequest, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[UserRecord] {
	return m.UpdateUserMetadataAfterRegistrationFn(request, ctx, forceLog)
}

func (m *UserGoWrapperMock) GetUsers(userIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]UserRecord] {
	return m.GetUsersFn(userIds, ctx, forceLog)
}

func (m *UserGoWrapperMock) GetUsersDetails(userIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]UserDetailRecord] {
	return m.GetUsersDetailFn(userIds, ctx, forceLog)
}

func (m *UserGoWrapperMock) GetPets(petIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]PetRecord] {
	return m.GetPetsFn(petIds, ctx, forceLog)
}

func (m *UserGoWrapperMock) GetPetsDetails(petIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]PetDetailRecord] {
	return m.GetPetsDetailFn(petIds, ctx, forceLog)
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

func (m *UserGoWrapperMock) AuthGuest(deviceId string, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[AuthGuestResp] {
	return m.AuthGuestFn(deviceId, apmTransaction, forceLog)
}

func (m *UserGoWrapperMock) GetBlockList(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[string][]int64] {
	return m.GetBlockListFn(userIds, apmTransaction, forceLog)
}

func (m *UserGoWrapperMock) GetUserBlock(blockedTo int64, blockedBy int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[UserBlockData] {
	return m.GetUserBlockFn(blockedTo, blockedBy, apmTransaction, forceLog)
}

func (m *UserGoWrapperMock) GetAllActiveBots(ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[GetAllActiveBotsResponse] {
	return m.GetAllActiveBotsFn(ctx, forceLog)
}

func (m *UserGoWrapperMock) GetConfigPropertiesInternal(properties []string, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[GetConfigPropertiesResponseChan] {
	return m.GetConfigPropertiesInternalFn(properties, ctx, forceLog)
}

func (m *UserGoWrapperMock) UpdateEmailMarketing(userId int64, emailMarketing null.String, emailMarketingVerified bool, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[any] {
	return m.UpdateEmailMarketingFn(userId, emailMarketing, emailMarketingVerified, ctx, forceLog)
}

func (m *UserGoWrapperMock) GenerateDeeplink(urlPath string, ctx context.Context,
	forceLog bool) chan wrappers.GenericResponseChan[GenerateDeeplinkResponse] {
	return m.GenerateDeeplinkFn(urlPath, ctx, forceLog)
}

func (m *UserGoWrapperMock) CreateExport(name string, exportType ExportType, filters interface{}, exportedBy int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[CreateExportResponse] {
	return m.CreateExportFn(name, exportType, filters, exportedBy, ctx, forceLog)
}

func (m *UserGoWrapperMock) FinalizeExport(exportId int64, file null.String, err error, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[FinalizeExportResponse] {
	return m.FinalizeExportFn(exportId, file, err, ctx, forceLog)
}

func (m *UserGoWrapperMock) GetGrandReferrerIds(ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[[]int64] {
	return m.GetGrandReferrerIdsFn(ctx, forceLog)
}

func (m *UserGoWrapperMock) SetSpotsUploadBanned(userId int64, banned bool, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[any] {
	return m.SetSpotsUploadBannedFn(userId, banned, ctx, forceLog)
}

func (m *UserGoWrapperMock) GetPetsSearch(keywords string, page, count int, ctx context.Context, forceLog bool) chan SearchPetDetailRecordResponseChan {
	return m.GetPetsSearchFn(keywords, page, count, ctx, forceLog)
}
func (m *UserGoWrapperMock) UpdatePetAlbum(petId int64, videoId string, userId int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[any] {
	return m.UpdatePetAlbumFn(petId, videoId, userId, ctx, forceLog)
}

func (m *UserGoWrapperMock) GetFriendListData(userId int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[GetFriendListDataResponse] {
	panic("implement me")
}

func (m *UserGoWrapperMock) GetUsersWithFollowers(userIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[GetFriendListDataResponse] {
	panic("implement me")
}

func GetMock() IUserGoWrapper { // for compiler errors
	return &UserGoWrapperMock{}
}
