package api

import (
	"context"
	"encoding/json"
	"github.com/digitalmonsters/configurator/pkg/configs"
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/callback"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"testing"
	"time"
)

func TestGetConfigs(t *testing.T) {
	a := apiApp{
		apiDef: map[string]swagger.ApiDescription{},
		service: &configs.ConfigServiceMock{
			AdminGetConfigsFn: func(db *gorm.DB, req configs.GetConfigRequest) (*configs.GetConfigResponse, error) {
				return &configs.GetConfigResponse{
					Items: []application.ConfigModel{
						application.ConfigModel{
							Key:         "test_key1",
							Value:       "45",
							Type:        application.ConfigTypeInteger,
							Description: "test key 1",
							AdminOnly:   false,
							CreatedAt:   time.Now().Add(-1 * time.Minute),
							UpdatedAt:   time.Now().Add(-1 * time.Minute),
							Category:    application.ConfigCategoryAd,
						},
						application.ConfigModel{
							Key:         "test_key2",
							Value:       "some text",
							Type:        application.ConfigTypeString,
							Description: "test key 2",
							AdminOnly:   false,
							CreatedAt:   time.Now().Add(-1 * time.Minute),
							UpdatedAt:   time.Now().Add(-1 * time.Minute),
							Category:    application.ConfigCategoryContent,
						},
					},
					TotalCount: 2,
				}, nil
			},
		},
	}

	var req = configs.GetConfigRequest{
		CreatedFrom: null.TimeFrom(time.Now().UTC().Add(-15 * time.Minute)),
		CreatedTo:   null.TimeFrom(time.Now().UTC()),
		UpdatedFrom: null.TimeFrom(time.Now().UTC().Add(-10 * time.Minute)),
		UpdatedTo:   null.TimeFrom(time.Now().UTC()),
		Limit:       10,
		Offset:      0,
	}
	js, err := json.Marshal(&req)
	if err != nil {
		t.Fatal(err)
	}
	resp, wrappedErr := a.getConfigs().GetFn()(js, router.MethodExecutionData{
		Context: context.TODO(),
		UserId:  int64(1),
	})
	if wrappedErr != nil {
		t.Fatal(wrappedErr.GetError())
	}
	var mappedResp = resp.(*configs.GetConfigResponse)
	assert.Equal(t, 2, len(mappedResp.Items))
	assert.Equal(t, int64(2), mappedResp.TotalCount)
}

func TestGetConfigLogs(t *testing.T) {
	a := apiApp{
		apiDef: map[string]swagger.ApiDescription{},
		service: &configs.ConfigServiceMock{
			AdminGetConfigLogsFn: func(db *gorm.DB, req configs.GetConfigLogsRequest, executionData router.MethodExecutionData) (*configs.GetConfigLogsResponse, error) {
				return &configs.GetConfigLogsResponse{
					Items: []configs.ConfigLogModel{
						configs.ConfigLogModel{
							Id:            int64(1),
							Key:           "test_key1",
							Value:         "45",
							CreatedAt:     time.Now().Add(-1 * time.Minute),
							UpdatedAt:     time.Now().Add(-1 * time.Minute),
							RelatedUserId: null.IntFrom(1),
						},
						configs.ConfigLogModel{
							Id:            int64(2),
							Key:           "test_key2",
							Value:         "some text",
							CreatedAt:     time.Now().Add(-1 * time.Minute),
							UpdatedAt:     time.Now().Add(-1 * time.Minute),
							RelatedUserId: null.IntFrom(1),
						},
					},
					TotalCount: 2,
				}, nil
			},
		},
	}
	var req = configs.GetConfigLogsRequest{
		Keys:        []string{"test_key1", "test_key2"},
		CreatedFrom: null.TimeFrom(time.Now().UTC().Add(-15 * time.Minute)),
		CreatedTo:   null.TimeFrom(time.Now().UTC()),
		UpdatedFrom: null.TimeFrom(time.Now().UTC().Add(-10 * time.Minute)),
		UpdatedTo:   null.TimeFrom(time.Now().UTC()),
		Limit:       10,
		Offset:      0,
	}
	js, err := json.Marshal(&req)
	if err != nil {
		t.Fatal(err)
	}
	resp, wrappedErr := a.getConfigLogs().GetFn()(js, router.MethodExecutionData{
		Context: context.TODO(),
		UserId:  int64(1),
	})
	if wrappedErr != nil {
		t.Fatal(wrappedErr.GetError())
	}
	var mappedResp = resp.(*configs.GetConfigLogsResponse)
	assert.Equal(t, 2, len(mappedResp.Items))
	assert.Equal(t, int64(2), mappedResp.TotalCount)
}

func TestUpsertConfig(t *testing.T) {
	var adminUserId = int64(1)
	a := apiApp{
		apiDef: map[string]swagger.ApiDescription{},
		service: &configs.ConfigServiceMock{
			AdminUpsertConfigFn: func(db *gorm.DB, req configs.UpsertConfigRequest, userId int64,
				publisher eventsourcing.Publisher[eventsourcing.ConfigEvent]) (*application.ConfigModel, []callback.Callback, error) {
				assert.Equal(t, userId, adminUserId)
				return &application.ConfigModel{
					Key:         req.Key,
					Value:       req.Value,
					Type:        req.Type,
					Description: req.Description,
					AdminOnly:   false,
					CreatedAt:   time.Now().UTC(),
					UpdatedAt:   time.Now().UTC(),
					Category:    req.Category,
				}, nil, nil
			},
		},
	}
	var req = configs.UpsertConfigRequest{
		Key:         "test_key3",
		Value:       "657",
		Type:        application.ConfigTypeInteger,
		Description: "test number",
		Category:    application.ConfigCategoryTokens,
	}
	js, err := json.Marshal(&req)
	if err != nil {
		t.Fatal(err)
	}
	resp, wrappedErr := a.upsertConfig().GetFn()(js, router.MethodExecutionData{
		Context: context.TODO(),
		UserId:  adminUserId,
	})
	if wrappedErr != nil {
		t.Fatal(wrappedErr.GetError())
	}
	var mappedResp = resp.(*application.ConfigModel)
	assert.Equal(t, req.Type, mappedResp.Type)
	assert.Equal(t, req.Value, mappedResp.Value)
	assert.Equal(t, req.Category, mappedResp.Category)
	assert.Equal(t, false, mappedResp.AdminOnly)
	assert.Equal(t, req.Description, mappedResp.Description)
}
