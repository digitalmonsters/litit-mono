package eventsourcing

import (
	"encoding/json"
	"fmt"
)

type EmailNotificationType string

const (
	EmailNotificationPasswordForgot EmailNotificationType = "email.password.forgot"
	EmailNotificationConfirmAddress EmailNotificationType = "email.confirm.address"
	EmailNotificationReferral       EmailNotificationType = "email.referral"
)

type EmailNotificationBasePayload struct {
	UserId int64 `json:"user_id"`
}

type EmailNotificationPasswordForgotPayload struct {
	EmailNotificationBasePayload
	Code int `json:"code"`
}

type EmailNotificationConfirmAddressPayload struct {
	EmailNotificationBasePayload
	Token     string `json:"token"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	Firstname string `json:"firstname"`
}

type EmailNotificationReferralPayload struct {
	EmailNotificationBasePayload
	UserName     string `json:"user_name"`
	ReferrerName string `json:"referrer_name"`
	NumReferrals int64  `json:"num_referrals"`
}

type EmailNotificationEventData struct {
	Type    EmailNotificationType `json:"type"`
	Payload json.RawMessage       `json:"payload"`
}

// GetPublishKey TODO: change after migrate from node
func (t EmailNotificationEventData) GetPublishKey() string {
	var payload EmailNotificationBasePayload

	if err := json.Unmarshal(t.Payload, &payload); err != nil {
		payload = EmailNotificationBasePayload{UserId: 0}
	}

	return fmt.Sprintf("%v", payload.UserId)
}
