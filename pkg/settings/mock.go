package settings

import (
	"context"
	"gorm.io/gorm"
)

type ServiceMock struct {
	GetPushSettingsFn        func(userId int64, ctx context.Context, db *gorm.DB) (map[string]bool, error)
	GetPushSettingsByAdminFn func(req GetPushSettingsByAdminRequest, ctx context.Context,
		db *gorm.DB) (map[string]GetPushSettingsByAdminItem, error)
	ChangePushSettingsFn      func(settings map[string]bool, userId int64, ctx context.Context) error
	IsPushNotificationMutedFn func(userId int64, templateId string, ctx context.Context) (bool, error)
}

func (s *ServiceMock) GetPushSettings(userId int64, ctx context.Context, db *gorm.DB) (map[string]bool, error) {
	return s.GetPushSettingsFn(userId, ctx, db)
}

func (s *ServiceMock) GetPushSettingsByAdmin(req GetPushSettingsByAdminRequest, ctx context.Context, db *gorm.DB) (map[string]GetPushSettingsByAdminItem, error) {
	return s.GetPushSettingsByAdmin(req, ctx, db)
}

func (s *ServiceMock) ChangePushSettings(settings map[string]bool, userId int64, ctx context.Context) error {
	return s.ChangePushSettingsFn(settings, userId, ctx)
}

func (s *ServiceMock) IsPushNotificationMuted(userId int64, templateId string, ctx context.Context) (bool, error) {
	return s.IsPushNotificationMutedFn(userId, templateId, ctx)
}

func GetMock() IService {
	return &ServiceMock{}
}
