package eventsourcing

import (
	"fmt"
	"time"

	"github.com/lib/pq"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
)

type PetEvent struct {
	PetId       int64          `json:"pet_id"`
	UserID      int64          `json:"user_id"`
	PetType     null.String    `json:"pet_type"`
	AvatarKey   null.String    `json:"avatar_key"`
	Name        null.String    `json:"name"`
	Breed       null.String    `json:"breed"`
	Gender      null.String    `json:"gender"`
	BirthDate   null.Time      `json:"birth_date"`
	Colors      pq.StringArray `gorm:"type:int[]" json:"colors"`
	Height      float32        `json:"height"`
	HeightUnit  null.String    `json:"height_unit"`
	Weight      float32        `json:"weight"`
	WeightUnit  null.String    `json:"weight_unit"`
	Description string         `json:"description"`
	Behaviors   pq.StringArray `gorm:"type:int[]" json:"behaviors"`
	Vaccines    pq.StringArray `gorm:"type:int[]" json:"vaccines"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at"`
	Deleted     bool           `json:"deleted"`
}

func (c PetEvent) GetPublishKey() string {
	return fmt.Sprint(c.PetId)
}
