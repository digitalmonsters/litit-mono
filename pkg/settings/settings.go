package settings

import (
	"context"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/gocql/gocql"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"time"
)

type IService interface {
	GetPushSettings(userId int64, ctx context.Context, db *gorm.DB) (map[string]bool, error)
	GetPushSettingsByAdmin(req GetPushSettingsByAdminRequest, ctx context.Context,
		db *gorm.DB) (map[string]GetPushSettingsByAdminItem, error)
	ChangePushSettings(settings map[string]bool, userId int64, ctx context.Context) error
	IsPushNotificationMuted(userId int64, templateId string, ctx context.Context) (bool, error)
}

type service struct {
	templatesCache *cache.Cache
}

func NewService() IService {
	return &service{
		templatesCache: cache.New(10*time.Minute, 10*time.Minute),
	}
}

func (s service) GetPushSettings(userId int64, ctx context.Context, db *gorm.DB) (map[string]bool, error) {
	session := database.GetScyllaSession()

	iter := session.Query("select template_id, muted from user_notifications_settings where "+
		"cluster_key = ? and user_id = ?", database.GetUserNotificationsSettingsClusterKey(userId), userId).
		WithContext(ctx).Iter()

	settingsMap := make(map[string]bool)

	var templateId string
	var muted bool

	for iter.Scan(&templateId, &muted) {
		settingsMap[templateId] = muted
	}

	if err := iter.Close(); err != nil {
		return nil, errors.WithStack(err)
	}

	templatesMap := make(map[string]database.RenderTemplate)

	for id, template := range s.templatesCache.Items() {
		templatesMap[id] = template.Object.(database.RenderTemplate)
	}

	if len(templatesMap) == 0 {
		var templates []database.RenderTemplate
		if err := db.Find(&templates).Error; err != nil {
			return nil, errors.WithStack(err)
		}

		for _, template := range templates {
			templatesMap[template.Id] = template
			s.templatesCache.Set(template.Id, template, 0)
		}
	}

	for id := range templatesMap {
		if _, ok := settingsMap[id]; !ok {
			settingsMap[id] = false
		}
	}

	return settingsMap, nil
}

func (s service) GetPushSettingsByAdmin(req GetPushSettingsByAdminRequest, ctx context.Context,
	db *gorm.DB) (map[string]GetPushSettingsByAdminItem, error) {
	session := database.GetScyllaSession()

	iter := session.Query("select template_id, muted from user_notifications_settings where "+
		"cluster_key = ? and user_id = ?", database.GetUserNotificationsSettingsClusterKey(req.UserId), req.UserId).
		WithContext(ctx).Iter()

	settingsMap := make(map[string]GetPushSettingsByAdminItem)

	var templateId string
	var muted bool

	for iter.Scan(&templateId, &muted) {
		settingsMap[templateId] = GetPushSettingsByAdminItem{Muted: muted}
	}

	if err := iter.Close(); err != nil {
		return nil, errors.WithStack(err)
	}

	templatesMap := make(map[string]database.RenderTemplate)

	for id, template := range s.templatesCache.Items() {
		templatesMap[id] = template.Object.(database.RenderTemplate)
	}

	if len(templatesMap) == 0 {
		var templates []database.RenderTemplate
		if err := db.Find(&templates).Error; err != nil {
			return nil, errors.WithStack(err)
		}

		for _, template := range templates {
			templatesMap[template.Id] = template
			s.templatesCache.Set(template.Id, template, 0)
		}
	}

	for id, template := range templatesMap {
		if v, ok := settingsMap[id]; !ok {
			settingsMap[id] = GetPushSettingsByAdminItem{
				RenderTemplate: template,
				Muted:          muted,
			}
		} else {
			v.RenderTemplate = template
		}
	}

	return settingsMap, nil
}

func (s service) ChangePushSettings(settings map[string]bool, userId int64, ctx context.Context) error {
	session := database.GetScyllaSession()

	var currentBatch = session.NewBatch(gocql.UnloggedBatch).WithContext(ctx)
	countStatements := 0

	for templateId, muted := range settings {
		currentBatch.Query("update user_notifications_settings set  muted = ? where cluster_key = ? and user_id = ? and template_id = ?",
			muted, database.GetUserNotificationsSettingsClusterKey(userId), userId, templateId)
		countStatements++

		if countStatements == 130 {
			if err := session.ExecuteBatch(currentBatch); err != nil {
				return errors.WithStack(err)
			}

			countStatements = 0
			currentBatch = session.NewBatch(gocql.UnloggedBatch).WithContext(ctx)
		}
	}

	if countStatements != 0 {
		if err := session.ExecuteBatch(currentBatch); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (s service) IsPushNotificationMuted(userId int64, templateId string, ctx context.Context) (bool, error) {
	session := database.GetScyllaSession()

	iter := session.Query("select muted from user_notifications_settings where "+
		"cluster_key = ? and user_id = ? and template_id = ?", database.GetUserNotificationsSettingsClusterKey(userId), userId, templateId).
		WithContext(ctx).Iter()

	var muted bool

	for iter.Scan(&muted) {
		break
	}

	if err := iter.Close(); err != nil {
		return false, errors.WithStack(err)
	}

	return muted, nil
}
