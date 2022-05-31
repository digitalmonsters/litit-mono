package sender

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/translation"
	"github.com/digitalmonsters/go-common/wrappers/notification_gateway"
	"github.com/digitalmonsters/go-common/wrappers/notification_handler"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/renderer"
	"github.com/digitalmonsters/notification-handler/pkg/token"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"strings"
)

type Sender struct {
	gateway notification_gateway.INotificationGatewayWrapper
}

func NewSender(gateway notification_gateway.INotificationGatewayWrapper) *Sender {
	return &Sender{
		gateway: gateway,
	}
}

func (s *Sender) SendTemplateToUser(channel notification_handler.NotificationChannel,
	title, body, headline string, renderingTemplate database.RenderTemplate, userId int64, renderingData map[string]string,
	customData map[string]interface{}, ctx context.Context) (interface{}, error) {
	db := database.GetDbWithContext(database.DbTypeReadonly, ctx)

	return s.sendPushTemplateMessageToUser(title, body, headline, renderingTemplate, userId, renderingData, customData, db, ctx)
}

func (s *Sender) SendCustomTemplateToUser(channel notification_handler.NotificationChannel, userId int64, pushType, kind,
	title, body, headline string, customData map[string]interface{}, ctx context.Context) (interface{}, error) {
	db := database.GetDbWithContext(database.DbTypeReadonly, ctx)

	return s.sendCustomPushTemplateMessageToUser(pushType, kind, title, body, headline, userId, customData, db, ctx)
}

func (s *Sender) SendEmail(msg []notification_gateway.SendEmailMessageRequest, ctx context.Context) error {
	return <-s.gateway.EnqueueEmail(msg, ctx)
}

func (s *Sender) sendPushTemplateMessageToUser(title, body, headline string,
	renderingTemplate database.RenderTemplate, userId int64, renderingData map[string]string,
	customData map[string]interface{}, db *gorm.DB, ctx context.Context) (interface{}, error) {
	userTokens, err := token.GetUserTokens(db, userId)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	if len(userTokens) == 0 {
		return nil, nil
	}
	sendResult := <-s.gateway.EnqueuePushForUser(s.preparePushEvents(userTokens, title, body,
		headline, renderingTemplate, fmt.Sprint(userId), renderingData, customData), ctx)

	return nil, sendResult
}

func (s *Sender) sendCustomPushTemplateMessageToUser(pushType, kind, title, body, headline string, userId int64, customData map[string]interface{},
	db *gorm.DB, ctx context.Context) (interface{}, error) {
	userTokens, err := token.GetUserTokens(db, userId)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	if len(userTokens) == 0 {
		return nil, nil
	}

	sendResult := <-s.gateway.EnqueuePushForUser(s.prepareCustomPushEvents(userTokens, pushType, kind, title, body, headline, fmt.Sprint(userId), customData), ctx)

	return nil, sendResult
}

func (s *Sender) preparePushEvents(tokens []database.Device, title string, body string, headline string, template database.RenderTemplate,
	key string, renderingData map[string]string, customData map[string]interface{}) []notification_gateway.SendPushRequest {
	mm := map[common.DeviceType]*notification_gateway.SendPushRequest{}

	extraData := map[string]string{
		"type":     template.Id,
		"kind":     template.Kind,
		"headline": headline,
	}
	if customData != nil {
		js, err := json.Marshal(&customData)
		if err != nil {
			log.Error().Str("push_type", template.Id).Str("push_kind", template.Kind).Err(err).Send()
		}
		extraData["custom_data"] = string(js)
	}

	for k, v := range renderingData {
		if _, ok := extraData[k]; !ok {
			extraData[k] = v
		}
	}

	for _, t := range tokens {
		if _, ok := mm[t.Platform]; !ok {
			req := notification_gateway.SendPushRequest{
				Tokens:     nil,
				DeviceType: t.Platform,
				Title:      title,
				Body:       body,
				ExtraData:  extraData,
				PublishKey: key,
			}

			mm[t.Platform] = &req
		}

		mm[t.Platform].Tokens = append(mm[t.Platform].Tokens, t.PushToken)
	}

	var resp []notification_gateway.SendPushRequest

	for _, v := range mm {
		resp = append(resp, *v)
	}

	return resp
}

func (s *Sender) prepareCustomPushEvents(tokens []database.Device, pushType, kind, title string, body string, headline string,
	key string, customData map[string]interface{}) []notification_gateway.SendPushRequest {
	mm := map[common.DeviceType]*notification_gateway.SendPushRequest{}

	var extraData = map[string]string{
		"type":     pushType,
		"kind":     kind,
		"headline": headline,
	}
	if customData != nil {
		js, err := json.Marshal(&customData)
		if err != nil {
			log.Error().Str("push_type", pushType).Str("push_kind", kind).Err(err).Send()
		}
		extraData["custom_data"] = string(js)
	}

	for _, t := range tokens {
		if _, ok := mm[t.Platform]; !ok {
			req := notification_gateway.SendPushRequest{
				Tokens:     nil,
				DeviceType: t.Platform,
				Title:      title,
				Body:       body,
				ExtraData:  extraData,
				PublishKey: key,
			}

			mm[t.Platform] = &req
		}

		mm[t.Platform].Tokens = append(mm[t.Platform].Tokens, t.PushToken)
	}

	var resp []notification_gateway.SendPushRequest

	for _, v := range mm {
		resp = append(resp, *v)
	}

	return resp
}

func (s *Sender) RenderTemplate(db *gorm.DB, templateName string, renderingData map[string]string,
	language translation.Language) (title string, body string, headline string, renderingTemplate database.RenderTemplate, err error) {
	var renderTemplate database.RenderTemplate

	if err := db.Where("id = ?", strings.ToLower(templateName)).Take(&renderTemplate).Error; err != nil {
		return "", "", "", renderTemplate, errors.WithStack(err)
	}

	title, body, headline, err = renderer.Render(renderTemplate, renderingData, language)

	return title, body, headline, renderTemplate, err
}
