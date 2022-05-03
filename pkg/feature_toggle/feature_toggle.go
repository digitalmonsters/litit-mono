package feature_toggle

import (
	"github.com/digitalmonsters/configurator/pkg/database"
	"github.com/pkg/errors"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"time"
)

func GetAllFeatureToggles(db *gorm.DB) (map[string]database.FeatureToggleConfig, error) {
	var toggles []database.FeatureToggle

	if err := db.Table("feature_toggles").Where("deleted_at is null").Find(&toggles).Error; err != nil {
		return nil, err
	}

	var togglesMap = make(map[string]database.FeatureToggleConfig)
	for _, t := range toggles {
		togglesMap[t.Key] = t.Value
	}
	return togglesMap, nil
}

func GetFeatureToggles(db *gorm.DB, request GetFeatureTogglesRequest) (*GetFeatureTogglesResponse, error) {
	var toggles []database.FeatureToggle

	var q = db.Table("feature_toggles")
	if len(request.Keys) > 0 {
		q = q.Where("key in ?", request.Keys)
	}
	if request.HideDeleted {
		q = q.Where("deleted_at is null")
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return nil, err
	}
	if err := q.Limit(request.Limit).Offset(request.Offset).Order("created_at desc").Find(&toggles).Error; err != nil {
		return nil, err
	}
	var models []FeatureToggleModel
	for _, f := range toggles {
		models = append(models, FeatureToggleModel{
			Id:        f.Id,
			Key:       f.Key,
			Value:     f.Value,
			CreatedAt: f.CreatedAt,
			UpdatedAt: f.UpdatedAt,
			DeletedAt: f.DeletedAt,
		})
	}
	return &GetFeatureTogglesResponse{
		Items:      models,
		TotalCount: count,
	}, nil
}

func CreateFeatureToggle(db *gorm.DB, request CreateFeatureToggleRequest) (*FeatureToggleModel, error) {
	var count int64
	if err := db.Model(database.FeatureToggle{}).Where("key = ?", request.Key).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("key already exists")
	}
	var dbModel = database.FeatureToggle{
		Key:   request.Key,
		Value: request.Value,
	}
	if err := db.Create(&dbModel).Error; err != nil {
		return nil, err
	}
	return &FeatureToggleModel{
		Id:        dbModel.Id,
		Key:       dbModel.Key,
		Value:     dbModel.Value,
		CreatedAt: dbModel.CreatedAt,
		UpdatedAt: dbModel.UpdatedAt,
		DeletedAt: dbModel.DeletedAt,
	}, nil
}

func UpdateFeatureToggle(db *gorm.DB, request UpdateFeatureToggleRequest) (*FeatureToggleModel, error) {
	var featureFlag database.FeatureToggle
	if err := db.Where("id = ?", request.Id).Find(&featureFlag).Error; err != nil {
		return nil, err
	}
	if featureFlag.Id == 0 {
		return nil, errors.New("feature flag not found")
	}
	featureFlag.Value = request.Value
	if err := db.Save(&featureFlag).Error; err != nil {
		return nil, err
	}
	return &FeatureToggleModel{
		Id:        featureFlag.Id,
		Key:       featureFlag.Key,
		Value:     featureFlag.Value,
		CreatedAt: featureFlag.CreatedAt,
		UpdatedAt: featureFlag.UpdatedAt,
		DeletedAt: featureFlag.DeletedAt,
	}, nil
}

func DeleteFeatureToggle(db *gorm.DB, request DeleteFeatureToggleRequest) error {
	var featureFlag database.FeatureToggle
	if err := db.Where("id = ?", request.Id).Find(&featureFlag).Error; err != nil {
		return err
	}
	if featureFlag.Id == 0 {
		return errors.New("feature flag not found")
	}
	featureFlag.DeletedAt = null.TimeFrom(time.Now().UTC())
	featureFlag.Value.Disable = true
	if err := db.Save(&featureFlag).Error; err != nil {
		return err
	}
	return nil
}

func CreateFeatureToggleEvents(db *gorm.DB, events []database.FeatureEvent) error {
	var featureEvents []database.FeatureToggleEvent
	for _, ev := range events {
		featureEvents = append(featureEvents, database.FeatureToggleEvent{FeatureToggleEvent: ev})
	}
	if err := db.Create(&featureEvents).Error; err != nil {
		return err
	}
	return nil
}
func ListFeatureToggleEvents(db *gorm.DB, limit, offset int) (*ListFeatureToggleEventsResponse, error) {
	var list []database.FeatureToggleEvent
	var count int64
	if err := db.Model(list).Count(&count).Error; err != nil {
		return nil, err
	}
	if err := db.Model(list).Order("created_at desc").Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, err
	}
	return &ListFeatureToggleEventsResponse{
		TotalCount: count,
		Items:      list,
	}, nil
}
