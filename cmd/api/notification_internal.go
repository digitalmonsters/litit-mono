package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/go-common/wrappers/notification_handler"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/firebase"
	"github.com/digitalmonsters/notification-handler/pkg/notification"
	"github.com/rs/zerolog/log"
)

func InitInternalNotificationApi(httpRouter *router.HttpRouter, apiDef map[string]swagger.ApiDescription, firebaseClient *firebase.FirebaseClient) error {
	getNotificationsReadCount := "GetNotificationsReadCount"
	disableUnregisteredTokens := "DisableUnregisteredTokens"
	createNotification := "CreateNotification"

	if err := httpRouter.GetRpcServiceEndpoint().RegisterRpcCommand(router.NewServiceCommand(getNotificationsReadCount,
		func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
			var req notification.GetNotificationsReadCountRequest

			if err := json.Unmarshal(request, &req); err != nil {
				return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
			}

			resp, err := notification.GetNotificationsReadCount(req, executionData.Context)
			if err != nil {
				return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
			}

			return resp, nil
		}, false)); err != nil {
		return err
	}

	apiDef[getNotificationsReadCount] = swagger.ApiDescription{
		Request: notification.GetNotificationsReadCountRequest{},
		Tags:    []string{"notification"},
	}

	if err := httpRouter.GetRpcServiceEndpoint().RegisterRpcCommand(router.NewServiceCommand(disableUnregisteredTokens,
		func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
			var req notification_handler.DisableUnregisteredTokensRequest

			if err := json.Unmarshal(request, &req); err != nil {
				return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
			}

			resp, err := notification.DisableUnregisteredTokens(req, database.GetDb(database.DbTypeMaster))
			if err != nil {
				return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
			}

			return resp, nil
		}, false)); err != nil {
		return err
	}

	apiDef[disableUnregisteredTokens] = swagger.ApiDescription{
		Request: notification_handler.DisableUnregisteredTokensRequest{},
		Tags:    []string{"notification"},
	}

	if err := httpRouter.GetRpcServiceEndpoint().RegisterRpcCommand(router.NewServiceCommand(createNotification,
		func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
			log.Info().Msg("Starting RPC command for createNotification")

			// Log the raw request payload
			log.Debug().Msgf("Raw request payload: %s", string(request))

			var req notification_handler.CreateNotificationRequest
			if err := json.Unmarshal(request, &req); err != nil {
				log.Error().Err(err).Msg("Failed to unmarshal request")
				return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
			}

			log.Info().Msgf("Parsed request: %+v", req)

			// Log database connection initialization
			log.Info().Msg("Initializing database connection")
			db := database.GetDb(database.DbTypeMaster)
			log.Info().Msg("Database connection initialized successfully")

			log.Info().Msg("Creating notification")
			resp, err := notification.CreateNotification(req, db)
			if err != nil {
				log.Error().Err(err).Msg("Failed to create notification")
				return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
			}

			log.Info().Msg("Fetching latest device information for user")
			deviceInfo, err := notification.GetLatestDeviceForUser(req.Notifications.UserID, db)
			if err != nil {
				log.Error().Err(err).Msg("Failed to fetch latest device information")
				return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
			}

			log.Info().Msgf("Fetched device info: %+v", deviceInfo)

			// Log notification sending conditions
			log.Info().Msgf("Notification type: %s, title: %s", req.Notifications.Type, req.Notifications.Title)
			if ((req.Notifications.Type == "push.content.successful-upload" || req.Notifications.Type == "push.intro.successful-upload") &&
				(req.Notifications.Title == "You got a reply" || req.Notifications.Title == "You got a video DM")) || req.Notifications.TriggerFireBase {
				log.Info().Msg("Sending push notification via Firebase")
				data := make(map[string]string)
				for k, v := range req.Notifications.CustomData {
					switch value := v.(type) {
					case string:
						data[k] = value
					case int:
						data[k] = fmt.Sprintf("%d", value)
					case float64:
						data[k] = fmt.Sprintf("%f", value)
					default:
						// Skip other types
					}
				}
				firebaseClient.SendNotification(context.Background(), deviceInfo.PushToken, req.Notifications.Title, req.Notifications.Message, req.Notifications.Type, data)
				log.Info().Msg("Push notification sent successfully")
			}

			go func() {
				gerr := notification.CreateInAppNotification(req, db)
				if gerr != nil {
					log.Error().Err(gerr).Msg("Failed to create in_app notification")
				}
			}()

			log.Info().Msg("Successfully created notification")
			return resp, nil
		}, false)); err != nil {
		log.Error().Err(err).Msg("Failed to register RPC command for createNotification")
		return err
	}

	apiDef[createNotification] = swagger.ApiDescription{
		Request: notification_handler.DisableUnregisteredTokensRequest{},
		Tags:    []string{"notification"},
	}
	return nil
}
