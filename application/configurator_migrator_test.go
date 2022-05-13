package application

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func TestHttpMigrator(t *testing.T) {
	port := "8090"

	var configsMap = map[string]MigrateConfigModel{
		"test key 1": {
			Key:            "test key 1",
			Value:          "50",
			Type:           ConfigTypeInteger,
			Description:    "test key 1 description",
			AdminOnly:      false,
			Category:       ConfigCategoryAd,
			ReleaseVersion: "v6.142",
		},
		"test key 2": {
			Key:            "test key 2",
			Value:          "some text",
			Type:           ConfigTypeString,
			Description:    "test key 2 description",
			AdminOnly:      true,
			Category:       ConfigCategoryTokens,
			ReleaseVersion: "v2.82",
		},
	}

	go func() {
		http.HandleFunc("/json/migrator", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			d, err := ioutil.ReadAll(req.Body)

			if err != nil {
				t.Error(err)
			}

			var r MigratorRequest

			if err := json.Unmarshal(d, &r); err != nil {
				t.Error(err)
			}
			var foundCounter = 0

			var resp = map[string]ConfigModel{}

			for key, val := range r.Configs {
				if configVal, ok := configsMap[key]; ok {
					foundCounter++
					assert.Equal(t, val.Key, configVal.Key)
					assert.Equal(t, val.Description, configVal.Description)
					assert.Equal(t, val.Value, configVal.Value)
					assert.Equal(t, val.ReleaseVersion, configVal.ReleaseVersion)
					assert.Equal(t, val.AdminOnly, configVal.AdminOnly)
					assert.Equal(t, val.Category, configVal.Category)
					assert.Equal(t, val.Type, configVal.Type)

					resp[key] = ConfigModel{
						Key:            val.Key,
						Value:          val.Value,
						Type:           val.Type,
						Description:    val.Description,
						AdminOnly:      val.AdminOnly,
						CreatedAt:      time.Now().UTC(),
						UpdatedAt:      time.Now().UTC(),
						Category:       val.Category,
						ReleaseVersion: val.ReleaseVersion,
					}
				}
			}
			assert.Equal(t, 2, foundCounter)
			js, err := json.Marshal(&resp)
			if err != nil {
				t.Error(err)
			}
			_, _ = w.Write(js)
		})

		_ = http.ListenAndServe(fmt.Sprintf("127.0.0.1:%v", port), nil)
	}()

	time.Sleep(100 * time.Millisecond)

	retr := NewHttpMigrator(fmt.Sprintf("http://127.0.0.1:%v/json/migrator", port))
	retr.SetMigratorMap(configsMap)
	val, err := retr.Migrate(context.TODO())

	assert.Nil(t, err)
	assert.Equal(t, 2, len(val))
}
