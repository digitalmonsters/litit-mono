package notification_handler

import (
	"context"
	"encoding/json"

	"github.com/digitalmonsters/go-common/translation"
	"github.com/digitalmonsters/go-common/wrappers/notification_gateway"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/notification-handler/configs"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/firebase"
	"github.com/digitalmonsters/notification-handler/pkg/sender"
	settingsPkg "github.com/digitalmonsters/notification-handler/pkg/settings"

	"github.com/rs/zerolog/log"
)

func (h *NotificationHandlerWrapper) CreatePushNotification(noti database.Notification, entityId int, relatedEntityId int,
	templateName string, language translation.Language, ctx context.Context) {
	cfg := configs.GetConfig()
	jsonStr, _ := json.Marshal(cfg.Firebase.ServiceAccountJSON)
	firebaseClient := firebase.Initialize(ctx, string(jsonStr))
	settingsService := settingsPkg.NewService()
	jobber, _ := configs.GetJobber(&cfg.Jobber)
	userGoWrapper := user_go.NewUserGoWrapper(cfg.Wrappers.UserGo)

	notificationSender := sender.NewSender(notification_gateway.NewNotificationGatewayWrapper(
		cfg.Wrappers.NotificationGateway), settingsService, jobber, userGoWrapper, firebaseClient)

	retryCount := 3
	shouldRetry, innerErr := notificationSender.PushNotification(noti, "", int64(entityId), int64(relatedEntityId), templateName, language, "default", ctx)
	if innerErr != nil {
		log.Error().Err(innerErr).Msg("Failed to send push notification")
	}

	if shouldRetry && retryCount > 0 {
		// Retry
		notificationSender.PushNotification(noti, "", int64(entityId), int64(relatedEntityId), templateName, language, "default", ctx)
		retryCount -= 1
	}
}
