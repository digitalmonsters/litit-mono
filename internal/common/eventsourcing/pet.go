package eventsourcing

import (
	"fmt"
	"time"

	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
)

type PetEvent struct {
	PetId              int64                 `json:"pet_id"`
	UserID             int64                 `json:"user_id"`
	PetType            null.String           `json:"pet_type"`
	AvatarKey          null.String           `json:"avatar_key"`
	Name               null.String           `json:"name"`
	Breed              null.String           `json:"breed"`
	Gender             null.String           `json:"gender"`
	BirthDate          null.Time             `json:"birth_date"`
	Colors             pq.StringArray        `gorm:"type:int[]" json:"colors"`
	Height             float32               `json:"height"`
	HeightUnit         null.String           `json:"height_unit"`
	Weight             float32               `json:"weight"`
	WeightUnit         null.String           `json:"weight_unit"`
	Description        string                `json:"description"`
	Behaviors          pq.StringArray        `gorm:"type:int[]" json:"behaviors"`
	Vaccines           pq.StringArray        `gorm:"type:int[]" json:"vaccines"`
	CreatedAt          time.Time             `json:"created_at"`
	UpdatedAt          time.Time             `json:"updated_at"`
	DeletedAt          gorm.DeletedAt        `json:"deleted_at"`
	Deleted            bool                  `json:"deleted"`
	Avatar             null.String           `json:"avatar"`
	Verified           bool                  `json:"verified"`
	Followers          int                   `json:"followers"`
	Following          int                   `json:"following"`
	Likes              int                   `json:"likes"`
	Uploads            int                   `json:"uploads"`
	Views              int                   `json:"views"`
	Shares             int                   `json:"shares"`
	Comments           int                   `json:"comments"`
	TotalPoints        decimal.Decimal       `json:"total_points"`
	CurrentPoints      decimal.Decimal       `json:"current_points"`
	CollectedPoints    decimal.Decimal       `json:"collected_points"`
	VaultPoints        decimal.Decimal       `json:"vault_points"`
	AllTimeVaultPoints decimal.Decimal       `json:"all_time_vault_points"`
	VideosCount        int                   `json:"videos_count"`
	CreatorStatus      user_go.CreatorStatus `json:"creator_status"`
	AdDisabled         bool                  `json:"ad_disabled"`
	BannedTill         null.Time             `json:"banned_till"`
	UploadBanned       bool                  `json:"upload_banned"`
	BaseChangeEvent
}

func (c PetEvent) GetPublishKey() string {
	return fmt.Sprint(c.PetId)
}
