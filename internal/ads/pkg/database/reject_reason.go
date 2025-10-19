package database

import (
	"gopkg.in/guregu/null.v4"
	"time"
)

type RejectReason struct {
	Id        int64     `json:"id"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
	DeletedAt null.Time `json:"deleted_at"`
}

func (RejectReason) TableName() string {
	return "reject_reasons"
}
