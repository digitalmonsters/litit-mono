package database

import (
	"github.com/lib/pq"
	"gorm.io/gorm"
	"time"
)

type Message struct {
	Id                 int64
	Title              string
	Description        string
	Countries          pq.StringArray `gorm:"type:text[]"`
	VerificationStatus VerificationStatus
	AgeFrom            int8
	AgeTo              int8
	PointsFrom         float64
	PointsTo           float64
	IsActive           bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeactivatedAt      *time.Time
	DeletedAt          gorm.DeletedAt
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
