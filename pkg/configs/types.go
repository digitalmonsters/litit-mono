package configs

import (
	"github.com/digitalmonsters/configurator/pkg/database"
	"gopkg.in/guregu/null.v4"
	"time"
)

type GetConfigRequest struct {
	Keys                []string                  `json:"keys"`
	Types               []database.ConfigType     `json:"types"`
	DescriptionContains null.String               `json:"description_contains"`
	AdminOnly           null.Bool                 `json:"admin_only"`
	Categories          []database.ConfigCategory `json:"categories"`
	CreatedFrom         null.Time                 `json:"created_from"`
	CreatedTo           null.Time                 `json:"created_to"`
	UpdatedFrom         null.Time                 `json:"updated_from"`
	UpdatedTo           null.Time                 `json:"updated_to"`
	Limit               int                       `json:"limit"`
	Offset              int                       `json:"offset"`
}
type ConfigModel struct {
	Key         string                  `json:"key"`
	Value       string                  `json:"value"`
	Type        database.ConfigType     `json:"type"`
	Description string                  `json:"description"`
	AdminOnly   bool                    `json:"admin_only"`
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at"`
	Category    database.ConfigCategory `json:"category"`
}
type GetConfigResponse struct {
	Items      []ConfigModel `json:"items"`
	TotalCount int64         `json:"total_count"`
}

type UpsertConfigRequest struct {
	Key         string                  `json:"key"`
	Value       string                  `json:"value"`
	Type        database.ConfigType     `json:"type"`
	Description string                  `json:"description"`
	AdminOnly   bool                    `json:"admin_only"`
	Category    database.ConfigCategory `json:"category"`
}

type GetConfigLogsRequest struct {
	Keys           []string  `json:"keys"`
	RelatedUserIds []int64   `json:"related_user_ids"`
	CreatedFrom    null.Time `json:"created_from"`
	CreatedTo      null.Time `json:"created_to"`
	UpdatedFrom    null.Time `json:"updated_from"`
	UpdatedTo      null.Time `json:"updated_to"`
	Limit          int       `json:"limit"`
	Offset         int       `json:"offset"`
}

type ConfigLogModel struct {
	Id            int64
	Key           string
	Value         string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	RelatedUserId int64
}

type GetConfigLogsResponse struct {
	Items      []ConfigLogModel `json:"items"`
	TotalCount int64            `json:"total_count"`
}

type ConfigEvent struct {
	Key   string
	Value string
}

func (c ConfigEvent) GetPublishKey() string {
	return c.Key
}
