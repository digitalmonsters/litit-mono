package message

import (
	"github.com/digitalmonsters/ads-manager/pkg/database"
	"github.com/lib/pq"
	"gopkg.in/guregu/null.v4"
)

type UpsertMessageAdminRequest struct {
	Items []adminMessage `json:"items"`
}

type adminMessage struct {
	Id                 null.Int             `json:"id"`
	Type               database.MessageType `json:"type"`
	Title              string               `json:"title"`
	Description        string               `json:"description"`
	Countries          pq.StringArray       `json:"countries"`
	VerificationStatus null.Int             `json:"verification_status"`
	AgeFrom            int8                 `json:"age_from"`
	AgeTo              int8                 `json:"age_to"`
	PointsFrom         float64              `json:"points_from"`
	PointsTo           float64              `json:"points_to"`
	IsActive           bool                 `json:"is_active"`
}

type DeleteMessagesBulkAdminRequest struct {
	Ids []int64 `json:"ids"`
}

type MessagesListAdminRequest struct {
	Keyword            null.String                  `json:"keyword"`
	Countries          []string                     `json:"countries"`
	VerificationStatus *database.VerificationStatus `json:"verification_status"`
	AgeFromFrom        int8                         `json:"age_from_from"`
	AgeFromTo          int8                         `json:"age_from_to"`
	AgeToFrom          int8                         `json:"age_to_from"`
	AgeToTo            int8                         `json:"age_to_to"`
	PointsFromFrom     float64                      `json:"points_from_from"`
	PointsFromTo       float64                      `json:"points_from_to"`
	PointsToFrom       float64                      `json:"points_to_from"`
	PointsToTo         float64                      `json:"points_to_to"`
	IsActive           null.Bool                    `json:"is_active"`
	Limit              int                          `json:"limit"`
	Offset             int                          `json:"offset"`
}

type MessagesListAdminResponse struct {
	Items      []database.Message `json:"items"`
	TotalCount int64              `json:"total_count"`
}

type GetMessageForUserRequest struct {
	UserId int64 `json:"user_id"`
}

type HardcodedUserResponse struct {
	Age                int     `json:"age"`
	VerificationStatus int     `json:"verification_status"`
	PointsCount        float64 `json:"points_count"`
	Country            string  `json:"country"`
}

type NotificationMessage struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}
