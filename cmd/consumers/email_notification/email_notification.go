package email_notification

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/wrappers/notification_gateway"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/notification-handler/configs"
	"github.com/digitalmonsters/notification-handler/pkg/database/scylla"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/digitalmonsters/notification-handler/pkg/utils"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
	"strconv"
)

func process(event newSendingEvent, ctx context.Context, notifySender sender.ISender, emailLinks configs.EmailLinks) (*kafka.Message, error) {
	var err error
	var template string
	var email string
	templateData := map[string]string{}
	var publishKey string

	apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "event_type", string(event.Type))

	switch event.Type {
	case eventsourcing.EmailNotificationPasswordForgot:
		var payload eventsourcing.EmailNotificationPasswordForgotPayload

		if err = json.Unmarshal(event.Payload, &payload); err != nil {
			return &event.Messages, err
		}

		apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "user_id", payload.UserId)

		var user *scylla.User
		user, err = utils.GetUser(payload.UserId, ctx)
		if err != nil {
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
		template = "forgot_password_link"
		publishKey = strconv.FormatInt(payload.UserId, 10)

		firstName, lastName := userRecord.GetFirstAndLastNameWithPrivacy()

		templateData["name"] = fmt.Sprintf("%v %v", firstName, lastName)
		templateData["code"] = strconv.Itoa(payload.Code)
	case eventsourcing.EmailNotificationConfirmAddress:
		var payload eventsourcing.EmailNotificationConfirmAddressPayload

		if err = json.Unmarshal(event.Payload, &payload); err != nil {
			return &event.Messages, err
		}

		apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "user_id", payload.UserId)

		email = payload.Email

		apm_helper.AddApmLabel(apm.TransactionFromContext(ctx), "email", email)

		template = "email_verify"
		publishKey = strconv.FormatInt(payload.UserId, 10)
		templateData["token"] = payload.Token
		templateData["username"] = payload.Username
		templateData["verifyMarketingSiteHost"] = emailLinks.VerifyHost
		templateData["verify_url_path"] = emailLinks.VerifyPath
		templateData["marketingSiteHost"] = emailLinks.MarketingSite
	case eventsourcing.EmailMarketingConfirmAddress:
		var payload eventsourcing.EmailMarketingNotificationConfirmAddressPayload

		if err = json.Unmarshal(event.Payload, &payload); err != nil {
			return &event.Messages, err
		}

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
			return &event.Messages, err
		}

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
		return &event.Messages, nil
	}

	templateDataMarshalled, err := json.Marshal(&templateData)
	if err != nil {
		return nil, err
	}

	emailRequest := notification_gateway.SendEmailMessageRequest{
		ToAddresses:  []string{email},
		Template:     template,
		TemplateData: string(templateDataMarshalled),
		PublishKey:   publishKey,
	}

	if err = notifySender.SendEmail([]notification_gateway.SendEmailMessageRequest{emailRequest}, ctx); err != nil {
		return nil, err
	}

	return &event.Messages, nil
}
