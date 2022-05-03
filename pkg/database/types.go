package database

import (
	"database/sql/driver"
	"encoding/json"
	"gopkg.in/guregu/null.v4"
	"time"
)

type FeatureToggle struct {
	Id        int64 `gorm:"primaryKey"`
	Key       string
	Value     FeatureToggleConfig `gorm:"type:jsonb"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt null.Time
}

func (FeatureToggle) TableName() string {
	return "feature_toggles"
}

type FeatureToggleConfig struct {
	Percentage  int              `json:"percentage"`
	True        interface{}      `json:"true"`
	False       interface{}      `json:"false"`
	Default     interface{}      `json:"default"`
	TrackEvents bool             `json:"trackEvents"`
	Rule        string           `json:"rule"`
	Version     float64          `json:"version"`
	Disable     bool             `json:"disable"`
	Rollout     *RolloutStrategy `json:"rollout"`
}

type RolloutStrategy struct {
	Progressive     *RolloutProgressive `json:"progressive"`
	Experimentation *ReleaseRamp        `json:"experimentation"`
	Scheduled       *RolloutScheduled   `json:"scheduled"`
}

type RolloutProgressive struct {
	Percentage  Percentage  `json:"percentage"`
	ReleaseRamp ReleaseRamp `json:"releaseRamp"`
}
type RolloutScheduled struct {
	Steps []Step `json:"steps"`
}

type Percentage struct {
	Initial int `json:"initial"`
	End     int `json:"end"`
}
type ReleaseRamp struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}
type Step struct {
	Date       time.Time `json:"date"`
	Rule       string    `json:"rule"`
	Percentage int       `json:"percentage"`
}

func (c *FeatureToggleConfig) Scan(input interface{}) error {
	return json.Unmarshal(input.([]byte), c)
}

func (c *FeatureToggleConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

type FeatureToggleEvent struct {
	Id                 int64
	FeatureToggleEvent FeatureEvent `gorm:"type:jsonb"`
	CreatedAt          time.Time
}

func (FeatureToggleEvent) TableName() string {
	return "feature_toggle_events"
}

type FeatureEvent struct {
	Kind         string      `json:"kind"`
	ContextKind  string      `json:"context_kind"`
	UserKey      string      `json:"user_key"`
	CreationDate int64       `json:"creation_date"`
	Key          string      `json:"key"`
	Variation    string      `json:"variation"`
	Val          interface{} `json:"value"`
	Default      bool        `json:"default"`
	Version      float64     `json:"version"`
}

func (f *FeatureEvent) Scan(input interface{}) error {
	return json.Unmarshal(input.([]byte), f)
}

func (f *FeatureEvent) Value() (driver.Value, error) {
	return json.Marshal(f)
}
