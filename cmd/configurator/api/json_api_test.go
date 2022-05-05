package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/configurator/pkg/configs"
	"github.com/digitalmonsters/configurator/pkg/database"
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
