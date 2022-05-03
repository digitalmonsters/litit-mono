package feature_toggle

import (
	"github.com/digitalmonsters/configurator/configs"
	"github.com/digitalmonsters/configurator/pkg/database"
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"os"
	"testing"
	"time"
)

var config configs.Settings
var gormDb *gorm.DB

func TestMain(m *testing.M) {
	config = configs.GetConfig()
	gormDb = database.GetDb(database.DbTypeMaster)
	os.Exit(m.Run())
}

func TestGetFeatureToggles(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresTables(config.MasterDb, []string{"public.feature_toggles"}, nil, t); err != nil {
		t.Fatal(err)
	}
	var createConfig = database.FeatureToggleConfig{
		Percentage:  1,
		True:        true,
		False:       false,
		Default:     false,
		TrackEvents: false,
		Rule:        "",
		Version:     1.1,
		Disable:     false,
		Rollout:     nil,
	}
	var key = "test_toggle_1"
	resp, err := CreateFeatureToggle(gormDb, CreateFeatureToggleRequest{
		Key:   key,
		Value: createConfig,
	})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, key, resp.Key)
	Compare(t, createConfig, resp.Value)

	var key2 = "test_toggle_2"
	createConfig2 := database.FeatureToggleConfig{
		Percentage: 11,
		True: struct {
			is_test bool
			count   float64
		}{
			is_test: true,
			count:   1.1,
		},
		False: struct {
			is_test    bool
			count      float64
			other_cond int
		}{
			is_test:    false,
			count:      1.34,
			other_cond: 14,
		},
		Default:     false,
		TrackEvents: false,
		Rule:        "",
		Version:     1.1,
		Disable:     false,
		Rollout: &database.RolloutStrategy{
			Progressive: &database.RolloutProgressive{
				Percentage: database.Percentage{
					Initial: 2,
					End:     12,
				},
				ReleaseRamp: database.ReleaseRamp{
					Start: time.Now().Add(-5 * time.Minute),
					End:   time.Now().Add(5 * time.Minute),
				},
			},
			Experimentation: &database.ReleaseRamp{
				Start: time.Now().Add(-7 * time.Minute),
				End:   time.Now().Add(17 * time.Minute),
			},
			Scheduled: &database.RolloutScheduled{Steps: []database.Step{
				database.Step{
					Date:       time.Now().Add(17 * time.Minute),
					Rule:       "test_rule",
					Percentage: 1,
				},
			}},
		},
	}

	resp2, err := CreateFeatureToggle(gormDb, CreateFeatureToggleRequest{
		Key:   key2,
		Value: createConfig2,
	})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, key2, resp2.Key)
	Compare(t, createConfig2, resp2.Value)

	var featureToggles []database.FeatureToggle
	if err := gormDb.Order("id desc").Find(&featureToggles).Error; err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 2, len(featureToggles))
	assert.Equal(t, key2, featureToggles[0].Key)
	Compare(t, createConfig2, featureToggles[0].Value)
	assert.Equal(t, key, featureToggles[1].Key)
	Compare(t, createConfig, featureToggles[1].Value)

	featuresResp, err := GetFeatureToggles(gormDb, GetFeatureTogglesRequest{
		Keys:        []string{key, key2},
		Limit:       10,
		Offset:      0,
		HideDeleted: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 2, len(featuresResp.Items))
	assert.Equal(t, int64(2), featuresResp.TotalCount)
	assert.Equal(t, key2, featureToggles[0].Key)
	Compare(t, createConfig2, featureToggles[0].Value)
	assert.Equal(t, key, featureToggles[1].Key)
	Compare(t, createConfig, featureToggles[1].Value)

	createConfig.Rollout = &database.RolloutStrategy{
		Progressive: &database.RolloutProgressive{
			Percentage: database.Percentage{
				Initial: 1,
				End:     20,
			},
			ReleaseRamp: database.ReleaseRamp{
				Start: time.Now().Add(-5 * time.Minute),
				End:   time.Now().Add(5 * time.Minute),
			},
		},
		Experimentation: nil,
		Scheduled:       nil,
	}
	resp3, err := UpdateFeatureToggle(gormDb, UpdateFeatureToggleRequest{
		Id:    resp.Id,
		Value: createConfig,
	})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, key, resp3.Key)
	Compare(t, createConfig, resp3.Value)

	err = DeleteFeatureToggle(gormDb, DeleteFeatureToggleRequest{Id: resp2.Id})
	if err != nil {
		t.Fatal(err)
	}

	featuresResp2, err := GetAllFeatureToggles(gormDb)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(featuresResp2))
	Compare(t, createConfig, featuresResp2[key])
}

func Compare(t *testing.T, old, new database.FeatureToggleConfig) {
	assert.Equal(t, old.Percentage, new.Percentage)
	assert.Equal(t, old.TrackEvents, new.TrackEvents)
	assert.Equal(t, old.Rule, new.Rule)
	assert.Equal(t, old.Version, new.Version)
	assert.Equal(t, old.Disable, new.Disable)
	if old.Rollout != nil {
		if old.Rollout.Experimentation != nil {
			assert.Equal(t, old.Rollout.Experimentation.End.Format(time.RFC3339), new.Rollout.Experimentation.End.Format(time.RFC3339))
			assert.Equal(t, old.Rollout.Experimentation.Start.Format(time.RFC3339), new.Rollout.Experimentation.Start.Format(time.RFC3339))
		}
		if old.Rollout.Progressive != nil {
			assert.Equal(t, old.Rollout.Progressive.ReleaseRamp.End.Format(time.RFC3339), new.Rollout.Progressive.ReleaseRamp.End.Format(time.RFC3339))
			assert.Equal(t, old.Rollout.Progressive.ReleaseRamp.Start.Format(time.RFC3339), new.Rollout.Progressive.ReleaseRamp.Start.Format(time.RFC3339))
			assert.Equal(t, old.Rollout.Progressive.Percentage.Initial, new.Rollout.Progressive.Percentage.Initial)
			assert.Equal(t, old.Rollout.Progressive.Percentage.End, new.Rollout.Progressive.Percentage.End)
		}
		if old.Rollout.Scheduled != nil {
			assert.Equal(t, len(old.Rollout.Scheduled.Steps), len(new.Rollout.Scheduled.Steps))
		}
	}
}

func TestListFeatureToggleEvents(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresTables(config.MasterDb, []string{"public.feature_toggle_events"}, nil, t); err != nil {
		t.Fatal(err)
	}
	var events []database.FeatureEvent
	for i := 0; i < 5; i++ {
		events = append(events, database.FeatureEvent{Version: 1, Default: false})
	}
	if err := CreateFeatureToggleEvents(gormDb, events); err != nil {
		t.Fatal(err)
	}
	resp, err := ListFeatureToggleEvents(gormDb, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 5, len(resp.Items))
	assert.Equal(t, int64(5), resp.TotalCount)
}
