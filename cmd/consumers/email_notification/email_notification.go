package email_notification

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/wrappers/notification_gateway"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/notification-handler/configs"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
	"strconv"
)

func process(event newSendingEvent, ctx context.Context, notifySender sender.ISender, userGoWrapper user_go.IUserGoWrapper,
	emailLinks configs.EmailLinks, apmTransaction *apm.Transaction) (*kafka.Message, error) {
	var err error
	var template string
	var email string
	templateData := map[string]string{}
	var publishKey string

	switch event.Type {
	case eventsourcing.EmailNotificationPasswordForgot:
		var payload eventsourcing.EmailNotificationPasswordForgotPayload

		if err = json.Unmarshal(event.Payload, &payload); err != nil {
			return &event.Messages, err
		}

		var userData user_go.UserRecord

		resp := <-userGoWrapper.GetUsers([]int64{payload.UserId}, apmTransaction, false)
		if resp.Error != nil {
			return nil, resp.Error.ToError()
		}

		var ok bool
		if userData, ok = resp.Items[payload.UserId]; !ok {
			return &event.Messages, errors.WithStack(errors.New("user not found")) // we should continue, no need to retry
		}

		email = userData.Email
		template = "forgot_password_link"
		publishKey = strconv.FormatInt(payload.UserId, 10)
		templateData["name"] = fmt.Sprintf("%v %v", userData.Firstname, userData.Lastname)
		templateData["code"] = strconv.Itoa(payload.Code)
	case eventsourcing.EmailNotificationConfirmAddress:
		var payload eventsourcing.EmailNotificationConfirmAddressPayload

		if err = json.Unmarshal(event.Payload, &payload); err != nil {
			return &event.Messages, err
		}

		email = payload.Email
		template = "email_verify"
		publishKey = strconv.FormatInt(payload.UserId, 10)
		templateData["token"] = payload.Token
		templateData["username"] = payload.Username
		templateData["verifyMarketingSiteHost"] = emailLinks.VerifyHost
		templateData["verify_url_path"] = emailLinks.VerifyPath
		templateData["marketingSiteHost"] = emailLinks.MarketingSite
	case eventsourcing.EmailNotificationReferral:
		var payload eventsourcing.EmailNotificationReferralPayload

		if err = json.Unmarshal(event.Payload, &payload); err != nil {
			return &event.Messages, err
		}

		var userData user_go.UserRecord

		resp := <-userGoWrapper.GetUsers([]int64{payload.UserId}, apmTransaction, false)
		if resp.Error != nil {
			return nil, resp.Error.ToError()
		}

		var ok bool
		if userData, ok = resp.Items[payload.UserId]; !ok {
			return &event.Messages, errors.WithStack(errors.New("user not found")) // we should continue, no need to retry
		}

		email = userData.Email
		template = "referal_sign_up"
		publishKey = strconv.FormatInt(payload.UserId, 10)
		templateData["name"] = userData.Firstname
		templateData["referred"] = payload.ReferrerName
		templateData["nth_referral"] = strconv.FormatInt(payload.NumReferrals, 10)
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
