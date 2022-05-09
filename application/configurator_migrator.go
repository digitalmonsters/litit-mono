package application

import (
	"context"
	"encoding/json"
	"github.com/digitalmonsters/go-common/http_client"
	"github.com/pkg/errors"
)

const (
	HttpMigratorDefaultUrl = "http://configurator/internal/json/migrator"
)

type Migrator interface {
	Migrate(ctx context.Context) (map[string]ConfigModel, error)
	SetMigratorMap(configsMap map[string]MigrateConfigModel)
}

func NewHttpMigrator(apiUrl string) Migrator {
	return &HttpMigrator{apiUrl: apiUrl}
}

type HttpMigrator struct {
	apiUrl      string
	migratorMap map[string]MigrateConfigModel
}

func (h *HttpMigrator) SetMigratorMap(configsMap map[string]MigrateConfigModel) {
	h.migratorMap = configsMap
}

func (h *HttpMigrator) Migrate(ctx context.Context) (map[string]ConfigModel, error) {
	if len(h.migratorMap) == 0 {
		return map[string]ConfigModel{}, nil
	}

	resp, err := http_client.DefaultHttpClient.NewRequest(ctx).WithForceLog().EnableDump().
		SetBody(MigratorRequest{
			Configs: h.migratorMap,
		}).
		Post(h.apiUrl)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	result := map[string]ConfigModel{}

	if err = json.Unmarshal(resp.Bytes(), &result); err != nil {
		return nil, errors.WithStack(err)
	}

	return result, nil
}
