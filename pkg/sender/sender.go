package sender

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/wrappers/notification_gateway"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/renderer"
	"github.com/digitalmonsters/notification-handler/pkg/token"
	"github.com/pkg/errors"
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

func (s *Sender) SendTemplateToUser(channel NotificationChannel,
	templateName string, userId int64, renderingData map[string]string,
	ctx context.Context) (interface{}, error) {
	db := database.GetDbWithContext(database.DbTypeReadonly, ctx)

	return s.sendPushTemplateMessageToUser(templateName, userId, renderingData, db, ctx)
}

func (s *Sender) SendCustomTemplateToUser(channel NotificationChannel, userId int64, title, body, headline string, ctx context.Context) (interface{}, error) {
	db := database.GetDbWithContext(database.DbTypeReadonly, ctx)

	return s.sendCustomPushTemplateMessageToUser(title, body, headline, userId, db, ctx)
}

func (s *Sender) sendPushTemplateMessageToUser(templateName string, userId int64, renderingData map[string]string,
	db *gorm.DB, ctx context.Context) (interface{}, error) {
	userTokens, err := token.GetUserTokens(db, userId)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	if len(userTokens) == 0 {
		return nil, nil
	}

	title, body, headline, renderingTemplate, err := s.renderTemplate(db, templateName, renderingData)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	sendResult := <-s.gateway.EnqueuePushForUser(s.preparePushEvents(userTokens, title, body, headline, renderingTemplate,
		fmt.Sprint(userId), renderingData), ctx)

	return nil, sendResult
}

func (s *Sender) sendCustomPushTemplateMessageToUser(title, body, headline string, userId int64,
	db *gorm.DB, ctx context.Context) (interface{}, error) {
	userTokens, err := token.GetUserTokens(db, userId)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	if len(userTokens) == 0 {
		return nil, nil
	}

	sendResult := <-s.gateway.EnqueuePushForUser(s.prepareCustomPushEvents(userTokens, title, body, headline, fmt.Sprint(userId)), ctx)

	return nil, sendResult
}

func (s *Sender) preparePushEvents(tokens []database.Device, title string, body string, headline string, template database.RenderTemplate,
	key string, renderingData map[string]string) []notification_gateway.SendPushRequest {
	mm := map[common.DeviceType]*notification_gateway.SendPushRequest{}

	extraData := map[string]string{
		"type":     template.Id,
		"kind":     template.Kind,
		"headline": headline,
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

func (s *Sender) prepareCustomPushEvents(tokens []database.Device, title string, body string, headline string, key string) []notification_gateway.SendPushRequest {
	mm := map[common.DeviceType]*notification_gateway.SendPushRequest{}

	for _, t := range tokens {
		if _, ok := mm[t.Platform]; !ok {
			req := notification_gateway.SendPushRequest{
				Tokens:     nil,
				DeviceType: t.Platform,
				Title:      title,
				Body:       body,
				ExtraData: map[string]string{
					"type":     "custom",
					"kind":     "popup",
					"headline": headline,
				},
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

func (s *Sender) renderTemplate(db *gorm.DB, templateName string,
	renderingData map[string]string) (title string, body string, headline string, renderingTemplate database.RenderTemplate, err error) {
	var renderTemplate database.RenderTemplate

	if err := db.Where("id = ?", strings.ToLower(templateName)).Take(&renderTemplate).Error; err != nil {
		return "", "", "", renderTemplate, errors.WithStack(err)
	}

	title, body, headline, err = renderer.Render(renderTemplate, renderingData)

	return title, body, headline, renderTemplate, err
}
