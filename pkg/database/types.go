package database

import (
	"github.com/lib/pq"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"time"
)

type Message struct {
	Id                 int64               `json:"id"`
	Type               MessageType         `json:"type"`
	Title              string              `json:"title"`
	Description        string              `json:"description"`
	Countries          pq.StringArray      `json:"countries" gorm:"type:text[]"`
	VerificationStatus *VerificationStatus `json:"verification_status"`
	AgeFrom            int8                `json:"age_from"`
	AgeTo              int8                `json:"age_to"`
	PointsFrom         float64             `json:"points_from"`
	PointsTo           float64             `json:"points_to"`
	IsActive           bool                `json:"is_active"`
	CreatedAt          time.Time           `json:"created_at"`
	UpdatedAt          time.Time           `json:"updated_at"`
	DeactivatedAt      *time.Time          `json:"deactivated_at"`
	DeletedAt          gorm.DeletedAt      `json:"deleted_at"`
}

func (Message) TableName() string {
	return "messages"
}

type MessageType int

const (
	MessageTypeMobile = MessageType(1)
	MessageTypeWeb    = MessageType(2)
)

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

type AdCampaign struct {
	Id             int64
	UserId         int64
	Name           string
	AdType         AdType
	Status         AdCampaignStatus
	ContentId      int64
	Link           null.String
	LinkButtonId   null.Int
	Country        null.String
	CreatedAt      time.Time
	StartedAt      null.Time
	EndedAt        null.Time
	DurationMin    uint
	Budget         uint
	Gender         null.String
	AgeFrom        uint
	AgeTo          uint
	RejectReasonId null.Int
}

func (AdCampaign) TableName() string {
	return "ad_campaigns"
}

type AdType int

const (
	AdTypeContent = AdType(1)
	AdTypeLink    = AdType(2)
)

type AdCampaignStatus int

const (
	AdCampaignStatusPending   = AdCampaignStatus(1)
	AdCampaignStatusModerated = AdCampaignStatus(2)
	AdCampaignStatusReject    = AdCampaignStatus(3)
	AdCampaignStatusActive    = AdCampaignStatus(4)
	AdCampaignStatusCompleted = AdCampaignStatus(5)
)
