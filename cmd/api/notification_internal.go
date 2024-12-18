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
			var req notification_handler.CreateNotificationRequest
			if err := json.Unmarshal(request, &req); err != nil {
				log.Error().Err(err).Msg("Failed to unmarshal request")
				return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
			}
			db := database.GetDb(database.DbTypeMaster)
			resp, err := notification.CreateNotification(req, db)
			if err != nil {
				return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
			}

			deviceInfo, err := notification.GetLatestDeviceForUser(req.Notifications.UserID, db)
			if err != nil {
				return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
			}
			fmt.Println(deviceInfo.PushToken)

			if req.Notifications.Type == "push.content.successful-upload" && req.Notifications.Title == "You got a reply" {
				firebaseClient.SendNotification(context.Background(), deviceInfo.PushToken, req.Notifications.Title, req.Notifications.Message, nil)
			}

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
