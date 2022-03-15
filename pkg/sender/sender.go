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
func (s *Sender) sendPushTemplateMessageToUser(templateName string, userId int64, renderingData map[string]string,
	db *gorm.DB, ctx context.Context) (interface{}, error) {
	userTokens, err := token.GetUserTokens(db, userId)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	if len(userTokens) == 0 {
		return nil, nil
	}

	title, body, renderingTemplate, err := s.renderTemplate(db, templateName, renderingData)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	sendResult := <-s.gateway.EnqueuePushForUser(s.preparePushEvents(userTokens, title, body, renderingTemplate,
		fmt.Sprint(userId)), ctx)

	return nil, sendResult
}

func (s *Sender) preparePushEvents(tokens []database.Device, title string, body string, template database.RenderTemplate,
	key string) []notification_gateway.SendPushRequest {
	mm := map[common.DeviceType]*notification_gateway.SendPushRequest{}

	for _, t := range tokens {
		if _, ok := mm[t.Platform]; !ok {
			req := notification_gateway.SendPushRequest{
				Tokens:     nil,
				DeviceType: t.Platform,
				Title:      title,
				Body:       body,
				ExtraData: map[string]string{
					"type": template.Id,
					"kind": template.Kind,
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
	renderingData map[string]string) (title string, body string, renderingTemplate database.RenderTemplate, err error) {
	var renderTemplate database.RenderTemplate

	if err := db.Where("id = ?", strings.ToLower(templateName)).Take(&renderTemplate).Error; err != nil {
		return "", "", renderTemplate, errors.WithStack(err)
	}

	title, body, err = renderer.Render(renderTemplate, renderingData)

	return title, body, renderTemplate, err
}
