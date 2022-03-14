package eventsourcing

import (
	"fmt"
	"github.com/lib/pq"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"time"
)

type CreatorStatus int

const (
	CreatorStatusNone     = CreatorStatus(0)
	CreatorStatusPending  = CreatorStatus(1)
	CreatorStatusRejected = CreatorStatus(2)
	CreatorStatusApproved = CreatorStatus(3)
)

type CreatorModel struct {
	Id           int64          `json:"id"`
	UserId       int64          `json:"user_id"`
	Status       CreatorStatus  `json:"status"`
	RejectReason null.String    `json:"reject_reason"`
	Url          string         `json:"url"`
	Categories   pq.Int64Array  `json:"categories"`
	CreatedAt    time.Time      `json:"created_at"`
	ApprovedAt   null.Time      `json:"approved_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at"`
}

func (c CreatorModel) GetPublishKey() string {
	return fmt.Sprint(c.UserId)
}
