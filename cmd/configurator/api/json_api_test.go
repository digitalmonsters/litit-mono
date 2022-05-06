package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/configurator/pkg/configs"
	"github.com/digitalmonsters/configurator/pkg/database"
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/callback"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/http_client"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"gorm.io/gorm"
	"testing"
	"time"
)

func TestJsonApi(t *testing.T) {
	a := apiApp{
		service: &configs.ConfigServiceMock{
			GetConfigsByIdsFn: func(db *gorm.DB, ids []string) ([]database.Config, error) {
				assert.Equal(t, 3, len(ids))

				return []database.Config{
					{
						Key:   "a",
						Value: "a1",
					},
					{
						Key:   "b",
						Value: "b1",
					},
					{
						Key:   "c",
						Value: "c1",
					},
				}, nil
			},
		},
	}

	go func() {
		if err := fasthttp.ListenAndServe(":8081", func(ctx *fasthttp.RequestCtx) {
			path := string(ctx.Request.URI().Path())

			fmt.Println(path)

			method, apiPath, handler := a.jsonApi()

			assert.Equal(t, "POST", method)
			assert.Equal(t, "/internal/json", apiPath)

			if path == apiPath && method == string(ctx.Method()) {
				handler(ctx)
			}
		}); err != nil {
			fmt.Println(err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	resp, err := http_client.DefaultHttpClient.NewRequest(context.TODO()).
		SetBody(configRequest{Items: []string{"a", "b", "c"}}).
		Post("http://localhost:8081/internal/json")

	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Fatal(fmt.Sprintf("unexpected status code %v. body %v", resp.StatusCode, resp.String()))
	}

	var results map[string]string

	if err := json.Unmarshal(resp.Bytes(), &results); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 3, len(results))

	assert.Equal(t, "a1", results["a"])
	assert.Equal(t, "b1", results["b"])
	assert.Equal(t, "c1", results["c"])
}

func TestJsonApiMigrator(t *testing.T) {
	a := apiApp{
		service: &configs.ConfigServiceMock{
			MigrateConfigsFn: func(db *gorm.DB, newConfigs map[string]application.MigrateConfigModel,
				publisher eventsourcing.Publisher[eventsourcing.ConfigEvent]) ([]application.ConfigModel, []callback.Callback, error) {
				var configsArr []application.ConfigModel

				for _, val := range newConfigs {
					configsArr = append(configsArr, application.ConfigModel{
						Key:            val.Key,
						Value:          val.Value,
						Type:           val.Type,
						Description:    val.Description,
						AdminOnly:      val.AdminOnly,
						CreatedAt:      time.Now().UTC(),
						UpdatedAt:      time.Now().UTC(),
						Category:       val.Category,
						ReleaseVersion: val.ReleaseVersion,
					})
				}
				return configsArr, nil, nil
			},
		},
	}
	go func() {
		if err := fasthttp.ListenAndServe(":8082", func(ctx *fasthttp.RequestCtx) {
			path := string(ctx.Request.URI().Path())

			fmt.Println(path)

			method, apiPath, handler := a.jsonMigratorApi()

			assert.Equal(t, "POST", method)
			assert.Equal(t, "/internal/json/migrator", apiPath)

			if path == apiPath && method == string(ctx.Method()) {
				handler(ctx)
			}
		}); err != nil {
			fmt.Println(err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	var reqValues = map[string]configs.UpsertConfigRequest{
		"test_key_1": configs.UpsertConfigRequest{
			Key:            "test_key_1",
			Value:          "50",
			Type:           application.ConfigTypeNumber,
			Description:    "test_key_1 description",
			AdminOnly:      false,
			Category:       application.ConfigCategoryContent,
			ReleaseVersion: "v1.5",
		},
		"test_key_2": configs.UpsertConfigRequest{
			Key:            "test_key_2",
			Value:          "some text",
			Type:           application.ConfigTypeString,
			Description:    "test_key_2 description",
			AdminOnly:      true,
			Category:       application.ConfigCategoryAd,
			ReleaseVersion: "v2.69",
		},
	}

	resp, err := http_client.DefaultHttpClient.NewRequest(context.TODO()).
		SetBody(migratorRequest{Configs: reqValues}).
		Post("http://localhost:8082/internal/json/migrator")

	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Fatal(fmt.Sprintf("unexpected status code %v. body %v", resp.StatusCode, resp.String()))
	}

	var resultsMap map[string]application.ConfigModel

	if err := json.Unmarshal(resp.Bytes(), &resultsMap); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 2, len(resultsMap))
	var foundCounter = 0
	for _, reqModel := range reqValues {
		if resultModel, ok := resultsMap[reqModel.Key]; ok {
			foundCounter++
			assert.Equal(t, reqModel.Key, resultModel.Key)
			assert.Equal(t, reqModel.Value, resultModel.Value)
			assert.Equal(t, reqModel.Category, resultModel.Category)
			assert.Equal(t, reqModel.Type, resultModel.Type)
			assert.Equal(t, reqModel.AdminOnly, resultModel.AdminOnly)
			assert.Equal(t, reqModel.ReleaseVersion, resultModel.ReleaseVersion)
			assert.Equal(t, reqModel.Description, resultModel.Description)
			assert.True(t, resultModel.CreatedAt.After(time.Now().UTC().Add(-5*time.Minute)))
			assert.True(t, resultModel.UpdatedAt.After(time.Now().UTC().Add(-5*time.Minute)))
		}
	}
	assert.Equal(t, 2, foundCounter)
}
