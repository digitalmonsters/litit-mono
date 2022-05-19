package configs

import (
	"github.com/digitalmonsters/go-common/application"
	"gopkg.in/guregu/null.v4"
	"time"
)

type GetConfigRequest struct {
	KeyLike            string `json:"key_like"`
	ReleaseVersionLike string `json:"release_version_like"`
	TypeLike           string `json:"type_like"`
	DescriptionLike    string `json:"description_contains"`
	CategoryLike       string `json:"category_like"`

	CreatedFrom null.Time `json:"created_from"`
	CreatedTo   null.Time `json:"created_to"`
	UpdatedFrom null.Time `json:"updated_from"`
	UpdatedTo   null.Time `json:"updated_to"`
	Limit       int       `json:"limit"`
	Offset      int       `json:"offset"`
}

type GetConfigResponse struct {
	Items      []application.ConfigModel `json:"items"`
	TotalCount int64                     `json:"total_count"`
}

type UpsertConfigRequest struct {
	Key         string                 `json:"key"`
	Value       string                 `json:"value"`
	Type        application.ConfigType `json:"type"`
	Description string                 `json:"description"`
	//AdminOnly      bool                       `json:"admin_only"`
	Category       application.ConfigCategory `json:"category"`
	ReleaseVersion string                     `json:"release_version"`
}

type GetConfigLogsRequest struct {
	Keys           []string    `json:"keys"`
	KeyContains    null.String `json:"key_contains"`
	RelatedUserIds []int64     `json:"related_user_ids"`
	CreatedFrom    null.Time   `json:"created_from"`
	CreatedTo      null.Time   `json:"created_to"`
	UpdatedFrom    null.Time   `json:"updated_from"`
	UpdatedTo      null.Time   `json:"updated_to"`
	Limit          int         `json:"limit"`
	Offset         int         `json:"offset"`
}

type ConfigLogModel struct {
	Id            int64
	Key           string
	Value         string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	RelatedUserId null.Int
}

type GetConfigLogsResponse struct {
	Items      []ConfigLogModel `json:"items"`
	TotalCount int64            `json:"total_count"`
}
