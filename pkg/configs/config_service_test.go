package configs

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/configurator/configs"
	"github.com/digitalmonsters/configurator/pkg/database"
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/wrappers/auth_go"
	"github.com/stretchr/testify/assert"
	"github.com/thoas/go-funk"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"os"
	"testing"
	"time"
)

var config configs.Settings
var gormDb *gorm.DB
var service *ConfigService

func TestMain(m *testing.M) {
	config = configs.GetConfig()
	gormDb = database.GetDb(database.DbTypeMaster)

	authWrapper := &auth_go.AuthGoWrapperMock{
		GetAdminsInfoByIdFn: func(userIds []int64, apmTx *apm.Transaction, forceLog bool) chan auth_go.GetAdminsInfoByIdResponseChan {
			var ch = make(chan auth_go.GetAdminsInfoByIdResponseChan, 2)
			go func() {
				var usersMap = make(map[int64]auth_go.AdminGeneralInfo)

				for _, userId := range userIds {
					usersMap[userId] = auth_go.AdminGeneralInfo{
						Name:  fmt.Sprint(userId),
						Email: fmt.Sprintf("test_email_%v", userId),
					}
				}
				ch <- auth_go.GetAdminsInfoByIdResponseChan{
					Items: usersMap,
				}
			}()
			return ch
		},
	}
	service = NewConfigService(authWrapper)

	os.Exit(m.Run())
}

func insertConfigs(t *testing.T, withConfigLog bool) ([]database.Config, []database.ConfigLog) {
	var configs = []database.Config{
		{
			Key:         "test_key1",
			Value:       "50",
			Type:        application.ConfigTypeInteger,
			Description: "test",
			AdminOnly:   false,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Category:    application.ConfigCategoryTokens,
		},
		{
			Key:         "test_key2",
			Value:       "test",
			Type:        application.ConfigTypeString,
			Description: "new description",
			AdminOnly:   false,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Category:    application.ConfigCategoryTokens,
		},
	}
	if err := gormDb.Create(&configs).Error; err != nil {
		t.Fatal(err)
	}
	if withConfigLog {
		var configLogs = []database.ConfigLog{
			database.ConfigLog{
				Key:           configs[0].Key,
				Value:         "165",
				RelatedUserId: null.IntFrom(1),
			},
			database.ConfigLog{
				Key:           configs[0].Key,
				Value:         "89",
				RelatedUserId: null.IntFrom(2),
			},
			database.ConfigLog{
				Key:           configs[0].Key,
				Value:         configs[0].Value,
				RelatedUserId: null.IntFrom(3),
			},
			database.ConfigLog{
				Key:           configs[1].Key,
				Value:         "some other text",
				RelatedUserId: null.IntFrom(1),
			},
			database.ConfigLog{
				Key:           configs[1].Key,
				Value:         "some other text 2",
				RelatedUserId: null.IntFrom(2),
			},
			database.ConfigLog{
				Key:           configs[1].Key,
				Value:         configs[1].Value,
				RelatedUserId: null.IntFrom(3),
			},
		}
		if err := gormDb.Create(&configLogs).Error; err != nil {
			t.Fatal(err)
		}
		return configs, configLogs
	}
	return configs, nil
}

func checkConfig(t *testing.T, old, new database.Config) {
	assert.Equal(t, old.Key, new.Key)
	assert.Equal(t, old.AdminOnly, new.AdminOnly)
	assert.Equal(t, old.Type, new.Type)
	assert.Equal(t, old.Category, new.Category)
	assert.Equal(t, old.Value, new.Value)
	assert.Equal(t, old.Description, new.Description)
	assert.Equal(t, old.UpdatedAt.UTC().Format(time.RFC3339), new.UpdatedAt.UTC().Format(time.RFC3339))
	assert.Equal(t, old.CreatedAt.UTC().Format(time.RFC3339), new.CreatedAt.UTC().Format(time.RFC3339))
}
func checkConfigModel(t *testing.T, old database.Config, new application.ConfigModel) {
	assert.Equal(t, old.Key, new.Key)
	assert.Equal(t, old.AdminOnly, new.AdminOnly)
	assert.Equal(t, old.Type, new.Type)
	assert.Equal(t, old.Category, new.Category)
	assert.Equal(t, old.Value, new.Value)
	assert.Equal(t, old.Description, new.Description)
	assert.Equal(t, old.UpdatedAt.UTC().Format(time.RFC3339), new.UpdatedAt.UTC().Format(time.RFC3339))
	assert.Equal(t, old.CreatedAt.UTC().Format(time.RFC3339), new.CreatedAt.UTC().Format(time.RFC3339))
}
func checkConfigMigrate(t *testing.T, old application.MigrateConfigModel, new application.ConfigModel) {
	assert.Equal(t, old.Key, new.Key)
	assert.Equal(t, old.AdminOnly, new.AdminOnly)
	assert.Equal(t, old.Type, new.Type)
	assert.Equal(t, old.Category, new.Category)
	assert.Equal(t, old.Value, new.Value)
	assert.Equal(t, old.Description, new.Description)
}
func checkConfigMigrateWithModel(t *testing.T, old application.MigrateConfigModel, new database.Config) {
	assert.Equal(t, old.Key, new.Key)
	assert.Equal(t, old.AdminOnly, new.AdminOnly)
	assert.Equal(t, old.Type, new.Type)
	assert.Equal(t, old.Category, new.Category)
	assert.Equal(t, old.Value, new.Value)
	assert.Equal(t, old.Description, new.Description)
}
func checkConfigLogs(t *testing.T, old database.ConfigLog, new ConfigLogModel, keyValues []string, relatedUserIds []int64) {
	assert.True(t, new.Id != 0)
	assert.Equal(t, old.Key, new.Key)
	assert.True(t, funk.Contains(relatedUserIds, new.RelatedUserId.Int64))
	assert.True(t, len(new.Username) > 0)
	assert.True(t, funk.Contains(keyValues, new.Value))
	assert.Equal(t, old.UpdatedAt.UTC().Format(time.RFC3339), new.UpdatedAt.UTC().Format(time.RFC3339))
	assert.Equal(t, old.CreatedAt.UTC().Format(time.RFC3339), new.CreatedAt.UTC().Format(time.RFC3339))
}

func TestConfigService_GetAllConfigs(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresAllTables(config.MasterDb, []string{"public.config"}, t); err != nil {
		t.Fatal(err)
	}
	configs, _ := insertConfigs(t, false)

	resp, err := service.GetAllConfigs(gormDb)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(configs), len(resp))
	var foundCounter int
	for _, r := range resp {
		for _, c := range configs {
			if r.Key == c.Key {
				foundCounter++
				checkConfig(t, c, r)
			}
		}
	}
	assert.Equal(t, 2, foundCounter)
}

func TestConfigService_GetConfigsByIds(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresAllTables(config.MasterDb, []string{"public.config"}, t); err != nil {
		t.Fatal(err)
	}
	configs, _ := insertConfigs(t, false)
	resp, err := service.GetConfigsByIds(gormDb, []string{configs[0].Key})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(resp))
	checkConfig(t, configs[0], resp[0])
}
func TestConfigService_AdminGetConfigs(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresAllTables(config.MasterDb, []string{"public.config"}, t); err != nil {
		t.Fatal(err)
	}
	configs, _ := insertConfigs(t, false)

	resp, err := service.AdminGetConfigs(gormDb, GetConfigRequest{
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(configs), len(resp.Items))
	assert.Equal(t, len(configs), int(resp.TotalCount))
	var foundCounter int
	for _, r := range resp.Items {
		for _, c := range configs {
			if r.Key == c.Key {
				foundCounter++
				checkConfigModel(t, c, r)
			}
		}
	}
	assert.Equal(t, 2, foundCounter)

	resp, err = service.AdminGetConfigs(gormDb, GetConfigRequest{
		KeyLike:     "test_key1",
		CreatedFrom: null.TimeFrom(time.Now().UTC().Add(-10 * time.Hour)),
		CreatedTo:   null.TimeFrom(time.Now().UTC().Add(1 * time.Hour)),
		UpdatedFrom: null.TimeFrom(time.Now().UTC().Add(-10 * time.Hour)),
		UpdatedTo:   null.TimeFrom(time.Now().UTC().Add(1 * time.Hour)),
		Limit:       10,
		Offset:      0,
	})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(resp.Items))
	assert.Equal(t, 1, int(resp.TotalCount))
	checkConfigModel(t, configs[0], resp.Items[0])
}
func TestConfigService_AdminGetConfigLogs(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresAllTables(config.MasterDb, []string{"public.config"}, t); err != nil {
		t.Fatal(err)
	}
	_, configLogs := insertConfigs(t, true)
	var values = make(map[string][]string)
	var relatedUsers = make(map[string][]int64)

	for _, c := range configLogs {
		values[c.Key] = append(values[c.Key], c.Value)
		relatedUsers[c.Key] = append(relatedUsers[c.Key], c.RelatedUserId.Int64)
	}

	resp, err := service.AdminGetConfigLogs(gormDb, GetConfigLogsRequest{
		Limit:  10,
		Offset: 0,
	}, router.MethodExecutionData{Context: context.TODO()})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(configLogs), len(resp.Items))
	assert.Equal(t, len(configLogs), int(resp.TotalCount))
	var foundCounterMap = make(map[int64]int)
	for _, r := range resp.Items {
		for _, c := range configLogs {
			if r.Key == c.Key {
				foundCounterMap[r.Id]++
				var keyValues = values[c.Key]
				var users = relatedUsers[c.Key]
				checkConfigLogs(t, c, r, keyValues, users)
			}
		}
	}
	assert.Equal(t, len(configLogs), len(foundCounterMap))

	resp, err = service.AdminGetConfigLogs(gormDb, GetConfigLogsRequest{
		Keys:        []string{configLogs[0].Key},
		KeyContains: null.StringFrom("test"),
		CreatedFrom: null.TimeFrom(time.Now().UTC().Add(-10 * time.Hour)),
		CreatedTo:   null.TimeFrom(time.Now().UTC().Add(1 * time.Hour)),
		UpdatedFrom: null.TimeFrom(time.Now().UTC().Add(-10 * time.Hour)),
		UpdatedTo:   null.TimeFrom(time.Now().UTC().Add(1 * time.Hour)),
		Limit:       10,
		Offset:      0,
	}, router.MethodExecutionData{Context: context.TODO()})
	if err != nil {
		t.Fatal(err)
	}
	foundCounterMap = make(map[int64]int)
	var firstKey = configLogs[0].Key
	var firstKeyValues = values[firstKey]
	var firstKeyUsers = relatedUsers[firstKey]
	for _, r := range resp.Items {
		for _, c := range configLogs {
			if r.Key == c.Key {
				foundCounterMap[r.Id]++
				checkConfigLogs(t, c, r, firstKeyValues, firstKeyUsers)
			}
		}
	}
	assert.Equal(t, 3, len(foundCounterMap))
}

func TestConfigService_ConfigLogs(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresAllTables(config.MasterDb, []string{"public.config"}, t); err != nil {
		t.Fatal(err)
	}
	var configs = []database.Config{
		{
			Key:            "test_key1",
			Value:          "50",
			Type:           application.ConfigTypeInteger,
			Description:    "test",
			AdminOnly:      false,
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
			Category:       application.ConfigCategoryTokens,
			ReleaseVersion: "1.2.1",
		},
	}
	if err := gormDb.Create(&configs).Error; err != nil {
		t.Fatal(err)
	}
	_, callbacks, err := service.AdminUpsertConfig(gormDb, UpsertConfigRequest{
		Key:            "test_key1",
		Value:          "123",
		Type:           application.ConfigTypeInteger,
		Description:    "test",
		Category:       application.ConfigCategoryTokens,
		ReleaseVersion: "1.2.1",
	}, 12, nil)

	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(callbacks))

	var logs []database.ConfigLog
	if err := gormDb.Find(&logs).Error; err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(logs))
	assert.Equal(t, configs[0].Key, logs[0].Key)
	assert.Equal(t, "123", logs[0].Value)
	assert.Equal(t, configs[0].Value, logs[0].OldValue)
	assert.Equal(t, int64(12), logs[0].RelatedUserId.Int64)

	resp, err := service.AdminGetConfigLogs(gormDb, GetConfigLogsRequest{}, router.MethodExecutionData{
		Context: context.TODO(),
	})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(resp.Items))
	assert.Equal(t, configs[0].Key, resp.Items[0].Key)
	assert.Equal(t, "123", resp.Items[0].Value)
	assert.Equal(t, configs[0].Value, resp.Items[0].OldValue)
	assert.Equal(t, int64(12), resp.Items[0].RelatedUserId.Int64)
	assert.Equal(t, fmt.Sprint(12), resp.Items[0].Username)
	assert.Equal(t, "test_email_12", resp.Items[0].Email)
}

func TestConfigService_AdminUpsertConfig(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresAllTables(config.MasterDb, []string{"public.config"}, t); err != nil {
		t.Fatal(err)
	}
	var publishedEvent eventsourcing.ConfigEvent

	var fn = func(ctx context.Context, messages ...eventsourcing.ConfigEvent) chan error {
		publishedEvent = messages[0]
		var errCh = make(chan error, 2)
		go func() {
			errCh <- nil
			defer close(errCh)
		}()
		return errCh
	}

	var publishMock = &eventsourcing.PublisherMock[eventsourcing.ConfigEvent]{
		PublishFn:          fn,
		PublishImmediateFn: fn,
	}
	if err := gormDb.Create(&database.Config{
		CreatedAt:      time.Time{},
		UpdatedAt:      time.Time{},
		Key:            "test key 3",
		Value:          "",
		Type:           "",
		Description:    "something other",
		AdminOnly:      false,
		Category:       "",
		ReleaseVersion: "",
	}).Error; err != nil {
		t.Fatal(err)
	}
	req := UpsertConfigRequest{
		Key:            "test key 3",
		Value:          "70",
		Type:           application.ConfigTypeInteger,
		Description:    "test number 3",
		Category:       application.ConfigCategoryContent,
		ReleaseVersion: "v1.2",
	}
	resp, callbacks, wrappedErr := service.AdminUpsertConfig(gormDb, req, int64(9), publishMock)
	if wrappedErr != nil {
		t.Fatal(wrappedErr)
	}
	for _, fn := range callbacks {
		if err := fn(context.TODO()); err != nil {
			t.Fatal(err)
		}
	}
	assert.Equal(t, resp.Key, publishedEvent.Key)
	assert.Equal(t, resp.Value, publishedEvent.Value)
	assert.Equal(t, req.Key, resp.Key)
	assert.Equal(t, false, resp.AdminOnly)
	assert.Equal(t, application.ConfigTypeInteger, resp.Type)
	assert.Equal(t, application.ConfigCategoryContent, resp.Category)
	assert.Equal(t, req.Value, resp.Value)
	assert.Equal(t, req.Description, resp.Description)

	assert.True(t, resp.CreatedAt.After(time.Now().UTC().Add(-1*time.Minute)))
	assert.True(t, resp.UpdatedAt.After(time.Now().UTC().Add(-1*time.Minute)))

	var cfg database.Config

	if err := gormDb.Where("key = ?", req.Key).Find(&cfg).Error; err != nil {
		t.Fatal(err)
	}
	checkConfigModel(t, cfg, *resp)

	req = UpsertConfigRequest{
		Key:            "test key 3",
		Value:          "test 71",
		Type:           application.ConfigTypeString,
		Description:    "test number 7",
		Category:       application.ConfigCategoryTokens,
		ReleaseVersion: "v1.1",
	}
	resp, callbacks, wrappedErr = service.AdminUpsertConfig(gormDb, req, int64(9), publishMock)
	if wrappedErr != nil {
		t.Fatal(wrappedErr)
	}
	for _, fn := range callbacks {
		if err := fn(context.TODO()); err != nil {
			t.Fatal(err)
		}
	}
	assert.Equal(t, resp.Key, publishedEvent.Key)
	assert.Equal(t, resp.Value, publishedEvent.Value)
	assert.Equal(t, resp.Key, req.Key)
	assert.Equal(t, false, resp.AdminOnly)
	assert.Equal(t, application.ConfigTypeInteger, resp.Type)
	assert.Equal(t, application.ConfigCategoryContent, resp.Category)
	assert.Equal(t, req.Value, resp.Value)
	assert.Equal(t, req.Description, resp.Description)
	assert.True(t, resp.CreatedAt.After(time.Now().UTC().Add(-1*time.Minute)))
	assert.True(t, resp.UpdatedAt.After(time.Now().UTC().Add(-1*time.Minute)))

	var cfg2 database.Config

	if err := gormDb.Where("key = ?", req.Key).Find(&cfg2).Error; err != nil {
		t.Fatal(err)
	}
	checkConfigModel(t, cfg2, *resp)

	configsResp, err := service.GetAllConfigs(gormDb)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(configsResp))
}

func TestConfigService_MigrateConfigs(t *testing.T) {
	if err := boilerplate_testing.FlushPostgresAllTables(config.MasterDb, []string{"public.config"}, t); err != nil {
		t.Fatal(err)
	}
	var publishedEvent []eventsourcing.ConfigEvent

	var fn = func(ctx context.Context, messages ...eventsourcing.ConfigEvent) chan error {
		publishedEvent = messages
		var errCh = make(chan error, 2)
		go func() {
			errCh <- nil
			defer close(errCh)
		}()
		return errCh
	}

	var publishMock = &eventsourcing.PublisherMock[eventsourcing.ConfigEvent]{
		PublishFn:          fn,
		PublishImmediateFn: fn,
	}
	var reqMap = make(map[string]application.MigrateConfigModel)
	reqMap["test key 1"] = application.MigrateConfigModel{
		Key:            "test key 1",
		Value:          "50",
		Type:           application.ConfigTypeInteger,
		Description:    "test key 1 description",
		AdminOnly:      false,
		Category:       application.ConfigCategoryAd,
		ReleaseVersion: "v1.32",
	}
	reqMap["test key 2"] = application.MigrateConfigModel{
		Key:            "test key 2",
		Value:          "some text",
		Type:           application.ConfigTypeString,
		Description:    "some text description",
		AdminOnly:      true,
		Category:       application.ConfigCategoryTokens,
		ReleaseVersion: "v12.3",
	}
	reqMap["test key 3"] = application.MigrateConfigModel{
		Key:            "test key 3",
		Value:          "{}",
		Type:           application.ConfigTypeObject,
		Description:    "object description",
		AdminOnly:      true,
		Category:       application.ConfigCategoryAd,
		ReleaseVersion: "v308.1",
	}
	resp, callbacks, wrappedErr := service.MigrateConfigs(gormDb, reqMap, publishMock)
	if wrappedErr != nil {
		t.Fatal(wrappedErr)
	}
	for _, fn := range callbacks {
		if err := fn(context.TODO()); err != nil {
			t.Fatal(err)
		}
	}
	var foundCounter int
	for _, val := range resp {
		if reqVal, ok := reqMap[val.Key]; ok {
			foundCounter++
			checkConfigMigrate(t, reqVal, val)
			assert.True(t, val.CreatedAt.After(time.Now().UTC().Add(-1*time.Minute)))
			assert.True(t, val.UpdatedAt.After(time.Now().UTC().Add(-1*time.Minute)))
		} else {
			t.Fatal("unknown config")
		}
	}
	assert.Equal(t, 3, foundCounter)
	var respMap = make(map[string]application.ConfigModel)
	for _, r := range resp {
		respMap[r.Key] = r
	}

	var cfg []database.Config

	if err := gormDb.Find(&cfg).Error; err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 3, len(cfg))
	foundCounter = 0
	for _, val := range cfg {
		if respVal, ok := respMap[val.Key]; ok {
			foundCounter++
			checkConfigModel(t, val, respVal)
			assert.True(t, val.CreatedAt.After(time.Now().UTC().Add(-1*time.Minute)))
			assert.True(t, val.UpdatedAt.After(time.Now().UTC().Add(-1*time.Minute)))
		} else {
			t.Fatal("unknown config")
		}
	}
	assert.Equal(t, 3, len(publishedEvent))
	assert.Equal(t, 3, foundCounter)
	foundCounter = 0
	for _, ev := range publishedEvent {
		if reqVal, ok := reqMap[ev.Key]; ok {
			foundCounter++
			assert.Equal(t, reqVal.Value, ev.Value)
		}
	}
	assert.Equal(t, 3, foundCounter)

	publishedEvent = []eventsourcing.ConfigEvent{}
	reqMap["test key 4"] = application.MigrateConfigModel{
		Key:            "test key 4",
		Value:          "123",
		Type:           application.ConfigTypeInteger,
		Description:    "test key 4 description",
		AdminOnly:      true,
		Category:       application.ConfigCategoryAd,
		ReleaseVersion: "v1.212",
	}
	resp, callbacks, wrappedErr = service.MigrateConfigs(gormDb, reqMap, publishMock)
	if wrappedErr != nil {
		t.Fatal(wrappedErr)
	}
	for _, fn := range callbacks {
		if err := fn(context.TODO()); err != nil {
			t.Fatal(err)
		}
	}
	foundCounter = 0
	for _, val := range resp {
		if reqVal, ok := reqMap[val.Key]; ok {
			foundCounter++
			assert.Equal(t, "test key 4", val.Key)
			checkConfigMigrate(t, reqVal, val)
			assert.True(t, val.CreatedAt.After(time.Now().UTC().Add(-1*time.Minute)))
			assert.True(t, val.UpdatedAt.After(time.Now().UTC().Add(-1*time.Minute)))
		} else {
			t.Fatal("unknown config")
		}
	}
	assert.Equal(t, 1, foundCounter)
	respMap = make(map[string]application.ConfigModel)
	for _, r := range resp {
		respMap[r.Key] = r
	}
	cfg = []database.Config{}

	if err := gormDb.Find(&cfg).Error; err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 4, len(cfg))
	foundCounter = 0
	for _, val := range cfg {
		if reqVal, ok := reqMap[val.Key]; ok {
			foundCounter++
			checkConfigMigrateWithModel(t, reqVal, val)
			assert.True(t, val.CreatedAt.After(time.Now().UTC().Add(-1*time.Minute)))
			assert.True(t, val.UpdatedAt.After(time.Now().UTC().Add(-1*time.Minute)))
		}
	}
	assert.Equal(t, 4, foundCounter)
	assert.Equal(t, 1, len(publishedEvent))
	foundCounter = 0
	for _, ev := range publishedEvent {
		if reqVal, ok := reqMap[ev.Key]; ok {
			foundCounter++
			assert.Equal(t, "test key 4", ev.Key)
			assert.Equal(t, reqVal.Value, ev.Value)
		}
	}
	assert.Equal(t, 1, foundCounter)
}
