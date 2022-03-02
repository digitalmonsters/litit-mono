package eventsourcing

import (
	"fmt"
	"github.com/shopspring/decimal"
	"time"
)

type UserBalanceChangeEvent struct {
	UserId int64 `json:"user_id"`

	TotalTokens decimal.Decimal `json:"total_tokens"`
	TotalPoints decimal.Decimal `json:"total_points"`

	CurrentPoints decimal.Decimal `json:"current_points"`
	CurrentTokens decimal.Decimal `json:"current_tokens"`
	CurrentRate   decimal.Decimal `json:"current_rate"`

	VaultPoints        decimal.Decimal `json:"vault_points"`
	AllTimeVaultPoints decimal.Decimal `json:"all_time_vault_points"`
}

func (u UserBalanceChangeEvent) GetPublishKey() string {
	return fmt.Sprintf("%v", u.UserId)
}

type PaidFeatureUpdateEvent struct {
	Id        int       `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Feature   string    `json:"feature"`
	UserId    int64     `json:"userId"`
}

func (u PaidFeatureUpdateEvent) GetPublishKey() string {
	return fmt.Sprintf("%v", u.UserId)
}
