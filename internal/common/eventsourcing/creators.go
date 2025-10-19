package eventsourcing

import (
	"fmt"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/lib/pq"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"time"
)

type CreatorModel struct {
	Id           int64                 `json:"id"`
	UserId       int64                 `json:"user_id"`
	Status       user_go.CreatorStatus `json:"status"`
	RejectReason null.String           `json:"reject_reason"`
	Url          string                `json:"url"`
	Categories   pq.Int64Array         `json:"categories"`
	CreatedAt    time.Time             `json:"created_at"`
	ApprovedAt   null.Time             `json:"approved_at"`
	DeletedAt    gorm.DeletedAt        `json:"deleted_at"`
}

func (c CreatorModel) GetPublishKey() string {
	return fmt.Sprint(c.UserId)
}

type MusicCreatorModel struct {
	Id           int64                 `json:"id"`
	UserId       int64                 `json:"user_id"`
	Status       user_go.CreatorStatus `json:"status"`
	RejectReason null.String           `json:"reject_reason"`
	LibraryUrl   string                `json:"link"`
	CreatedAt    time.Time             `json:"created_at"`
	ApprovedAt   null.Time             `json:"approved_at"`
	DeletedAt    gorm.DeletedAt        `json:"deleted_at"`
}

func (c MusicCreatorModel) GetPublishKey() string {
	return fmt.Sprint(c.UserId)
}
