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
	Id                 null.Int                    `json:"id"`
	Title              string                      `json:"title"`
	Description        string                      `json:"description"`
	Countries          pq.StringArray              `json:"countries"`
	VerificationStatus database.VerificationStatus `json:"verification_status"`
	AgeFrom            int8                        `json:"age_from"`
	AgeTo              int8                        `json:"age_to"`
	PointsFrom         float64                     `json:"points_from"`
	PointsTo           float64                     `json:"points_to"`
	IsActive           bool                        `json:"is_active"`
}

type DeleteMessagesBulkAdminRequest struct {
	Ids []int64 `json:"ids"`
}

type MessagesListAdminRequest struct {
	Keyword            null.String                 `json:"keyword"`
	Countries          []string                    `json:"countries"`
	VerificationStatus database.VerificationStatus `json:"verification_status"`
	AgeFrom            int8                        `json:"age_from"`
	AgeTo              int8                        `json:"age_to"`
	PointsFrom         float64                     `json:"points_from"`
	PointsTo           float64                     `json:"points_to"`
	IsActive           bool                        `json:"is_active"`
	Limit              int                         `json:"limit"`
	Offset             int                         `json:"offset"`
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
