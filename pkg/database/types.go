package database

import (
	"github.com/lib/pq"
	"gorm.io/gorm"
	"time"
)

type Message struct {
	Id                 int64              `json:"id"`
	Title              string             `json:"title"`
	Description        string             `json:"description"`
	Countries          pq.StringArray     `json:"countries" gorm:"type:text[]"`
	VerificationStatus VerificationStatus `json:"verification_status"`
	AgeFrom            int8               `json:"age_from"`
	AgeTo              int8               `json:"age_to"`
	PointsFrom         float64            `json:"points_from"`
	PointsTo           float64            `json:"points_to"`
	IsActive           bool               `json:"is_active"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
	DeactivatedAt      *time.Time         `json:"deactivated_at"`
	DeletedAt          gorm.DeletedAt     `json:"deleted_at"`
}

func (Message) TableName() string {
	return "messages"
}

type VerificationStatus int8

const (
	VerificationStatusNone     = VerificationStatus(0)
	VerificationStatusPending  = VerificationStatus(1)
	VerificationStatusVerified = VerificationStatus(2)
	VerificationStatusRejected = VerificationStatus(3)
)

func VerificationStatusFromString(status string) VerificationStatus {
	switch status {
	case "pending":
		return VerificationStatusPending
	case "verified":
		return VerificationStatusVerified
	case "rejected":
		return VerificationStatusRejected
	default:
		return VerificationStatusNone
	}
}
