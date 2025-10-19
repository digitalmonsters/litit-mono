package email_notification

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/wrappers/notification_gateway"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/notification-handler/configs"
	"github.com/digitalmonsters/notification-handler/pkg/database/scylla"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/digitalmonsters/notification-handler/pkg/utils"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
)

func process(event newSendingEvent, ctx context.Context, notifySender sender.ISender, emailLinks configs.EmailLinks) (*kafka.Message, error) {
	logger := log.Ctx(ctx).With().Str("event_type", string(event.Type)).Logger()
	var err error
	var template string
	var email string
	templateData := map[string]string{}
	var publishKey string
	var subject string
	var textBody string

	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "event_type", string(event.Type))

	logger.Info().Msgf("event.Type : %v", event.Type)

	switch event.Type {
	case eventsourcing.EmailNotificationPasswordForgot:
		var payload eventsourcing.EmailNotificationPasswordForgotPayload

		if err = json.Unmarshal(event.Payload, &payload); err != nil {
			logger.Error().Err(err).Msg("Failed to unmarshal EmailNotificationPasswordForgot payload")
			return &event.Messages, err
		}

		logger = logger.With().Str("user_id", strconv.FormatInt(payload.UserId, 10)).Logger()
		apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "user_id", payload.UserId)

		var user *scylla.User
		user, err = utils.GetUser(payload.UserId, ctx)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to get user")
			return nil, errors.WithStack(err)
		}

		userRecord := user_go.UserRecord{
			UserId:            user.UserId,
			Username:          user.Username,
			Firstname:         user.Firstname,
			Lastname:          user.Lastname,
			NamePrivacyStatus: user.NamePrivacyStatus,
		}

		email = user.Email
		publishKey = strconv.FormatInt(payload.UserId, 10)

		firstName, lastName := userRecord.GetFirstAndLastNameWithPrivacy()

		subject = "Password Reset Request"
		textBody = fmt.Sprintf("Hello %s %s,\n\nWe received a request to reset your password. Your reset code is: %d.\n\nIf you didn't request this, please ignore this email.", firstName, lastName, payload.Code)

	case eventsourcing.EmailNotificationConfirmAddress:
		var payload eventsourcing.EmailNotificationConfirmAddressPayload

		if err = json.Unmarshal(event.Payload, &payload); err != nil {
			logger.Error().Err(err).Msg("Failed to unmarshal EmailNotificationConfirmAddress payload")
			return &event.Messages, err
		}

		logger = logger.With().Str("user_id", strconv.FormatInt(payload.UserId, 10)).Logger()
		apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "user_id", payload.UserId)

		email = payload.Email
		apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "email", email)
		publishKey = strconv.FormatInt(payload.UserId, 10)

		// Prepare a well-formatted text body
		subject = "Email Address Confirmation"
		textBody = fmt.Sprintf(
			"Hello %s,\n\nThank you for registering! Please confirm your email address by entering the following confirmation code: %s.\n\nIf you didn't request this, please ignore this email.\n\nBest regards,\n Litit",
			payload.Firstname, payload.Token,
		)

	case eventsourcing.EmailMarketingConfirmAddress:
		var payload eventsourcing.EmailMarketingNotificationConfirmAddressPayload

		if err = json.Unmarshal(event.Payload, &payload); err != nil {
			logger.Error().Err(err).Msg("Failed to unmarshal EmailMarketingConfirmAddress payload")
			return &event.Messages, err
		}

		logger = logger.With().Str("user_id", strconv.FormatInt(payload.UserId, 10)).Logger()
		apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "user_id", payload.UserId)

		email = payload.Email

		apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "email", email)

		template = "email_marketing_verify"
		publishKey = strconv.FormatInt(payload.UserId, 10)
		templateData["token"] = payload.Token
		templateData["username"] = payload.Username
		templateData["verifymarketingsitehost"] = emailLinks.VerifyHost
		templateData["verify_url_path"] = emailLinks.VerifyEmailMarketingPath
		templateData["reward_points"] = payload.RewardPoints.String()
	case eventsourcing.EmailGuestTempInfo:
		var payload eventsourcing.EmailNotificationTempGuestInfoPayload

		if err = json.Unmarshal(event.Payload, &payload); err != nil {
			logger.Error().Err(err).Msg("Failed to unmarshal EmailGuestTempInfo payload")
			return &event.Messages, err
		}

		logger = logger.With().Str("user_id", strconv.FormatInt(payload.UserId, 10)).Logger()
		apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "user_id", payload.UserId)

		email = payload.Email

		apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "email", email)

		template = "email_temp_guest_info"
		publishKey = strconv.FormatInt(payload.UserId, 10)
		templateData["username"] = payload.Username
		templateData["deeplink"] = payload.DeepLink
	//case eventsourcing.EmailNotificationReferral:
	//	var payload eventsourcing.EmailNotificationReferralPayload
	//
	//	if err = json.Unmarshal(event.Payload, &payload); err != nil {
	//		return &event.Messages, err
	//	}
	//
	//	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx, "user_id", payload.UserId)
	//
	//	var userData user_go.UserRecord
	//
	//	resp := <-userGoWrapper.GetUsers([]int64{payload.UserId}, ctx, false)
	//	if resp.Error != nil {
	//		return nil, resp.Error.ToError()
	//	}
	//
	//	var ok bool
	//	if userData, ok = resp.Response[payload.UserId]; !ok {
	//		return &event.Messages, errors.WithStack(errors.New("user not found")) // we should continue, no need to retry
	//	}
	//
	//	email = userData.Email
	//	template = "referal_sign_up"
	//	publishKey = strconv.FormatInt(payload.UserId, 10)
	//
	//	firstName, _ := userData.GetFirstAndLastNameWithPrivacy()
	//
	//	templateData["name"] = firstName
	//	templateData["referred"] = payload.UserName
	//	templateData["nth_referral"] = strconv.FormatInt(payload.NumReferrals, 10)
	default:
		logger.Warn().Msg("Unknown event type")
		return &event.Messages, nil
	}

	templateDataMarshalled, err := json.Marshal(&templateData)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to marshal template data")
		return nil, err
	}

	emailRequest := notification_gateway.SendEmailMessageRequest{
		ToAddresses:  []string{email},
		Template:     template,
		TemplateData: string(templateDataMarshalled),
		PublishKey:   publishKey,
		Subject:      subject,
		TextBody:     textBody,
	}

	if err = notifySender.SendEmail([]notification_gateway.SendEmailMessageRequest{emailRequest}, ctx); err != nil {
		logger.Error().Err(err).Msg("Failed to send email")
		return nil, err
	}

	logger.Info().Msg("Email sent successfully")
	return &event.Messages, nil
}
