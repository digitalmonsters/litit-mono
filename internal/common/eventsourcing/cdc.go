package eventsourcing

import (
	"encoding/json"
	"fmt"
	"gopkg.in/guregu/null.v4"
	"time"
)

type CdcValueTime struct {
	Value CdcTimestamp `json:"value"`
}

type CdcValueBool struct {
	Value bool `json:"value"`
}

type CdcValueNullInt struct {
	Value null.Int `json:"value"`
}

type CdcValueInt struct {
	Value int `json:"value"`
}

type CdcValueInt64 struct {
	Value int64 `json:"value"`
}

type CdcValueString struct {
	Value string `json:"value"`
}

type CdcValueNullString struct {
	Value null.String `json:"value"`
}

type CdcTimestamp struct {
	time.Time
}

func (p *CdcTimestamp) UnmarshalJSON(bytes []byte) error {
	var raw int64
	err := json.Unmarshal(bytes, &raw)

	if err != nil {
		return err
	}

	p.Time = time.Unix(raw/1000, 0)

	return nil
}

func (p CdcTimestamp) MarshalJSON() ([]byte, error) {
	if p.Time.IsZero() {
		return nil, nil
	}

	return []byte(fmt.Sprintf("%v", p.Time.UnixMilli())), nil
}
