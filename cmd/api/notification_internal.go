package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/wrappers/notification_handler"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/firebase"
	"github.com/digitalmonsters/notification-handler/pkg/mail"
	"github.com/digitalmonsters/notification-handler/pkg/notification"
	"github.com/rs/zerolog/log"
)

func InitInternalNotificationApi(httpRouter *router.HttpRouter, firebaseClient *firebase.FirebaseClient, userGoWrapper user_go.IUserGoWrapper, mailSvc mail.IEmailService) error {
	getNotificationsReadCount := "GetNotificationsReadCount"
	disableUnregisteredTokens := "DisableUnregisteredTokens"
	createNotification := "CreateNotification"
	deleteNotificationByIntroID := "DeleteNotificationByIntroID"
	sendCustomEmail := "SendCustomEmail"
	sendCustomEmailWithGenericHTML := "SendCustomEmailWithGenericHTML"

	if err := httpRouter.GetRpcServiceEndpoint().RegisterRpcCommand(router.NewServiceCommand(deleteNotificationByIntroID,
		func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
			log.Info().Msg("Starting RPC command for DeleteNotificationByIntroID")

			var req notification.DeleteNotificationByIntroIDRequest

			if err := json.Unmarshal(request, &req); err != nil {
				return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
			}

			db := database.GetDb(database.DbTypeMaster)

			log.Info().Msg("Deleting notifications for given Intro ")

			err := notification.DeleteNotificationByIntroIDAndType(executionData.Context, db, req.IntroID)
			if err != nil {
				return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
			}
			log.Info().Msg("Successfully Deleted the notifications for given Intro ")

			return notification_handler.CreateNotificationResponse{
				Status: true,
			}, nil
		}, false)); err != nil {
		return err
	}

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

			if req.Notifications.Type == "push.profile.unfollowing" {
				log.Info().Int64("user_id", int64(req.Notifications.UserID)).Msg("Attempting to delete notification for unfollowing user")

				err := notification.DeleteUnFollowNotification(context.Background(), req, db)
				if err != nil {
					log.Error().Err(err).Int64("user_id", int64(req.Notifications.UserID)).Msg("Failed to delete notification for unfollowing user")
					return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
				}

				log.Info().Int64("user_id", int64(req.Notifications.UserID)).Msg("Successfully deleted notification for unfollowing user")
				return notification_handler.CreateNotificationResponse{Status: true}, nil
			}

			log.Info().Msg("Creating notification")
			resp, err := notification.CreateNotification(context.Background(), req, db)
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
						data[k] = fmt.Sprintf("%.0f", value)
					default:
						// Skip other types
					}
				}
				if req.Notifications.RelatedUserID != nil {
					resp := <-userGoWrapper.GetUsersDetails([]int64{int64(*req.Notifications.RelatedUserID)}, context.Background(), false)
					userResp, hasUser := (resp.Response)[int64(*req.Notifications.RelatedUserID)]
					if !hasUser {
						log.Info().Msg("Sending push notification via Firebase, not user found for RelatedUserID")
					}
					data["avatar_url"] = userResp.Avatar.String
				}
				log.Warn().Msgf("req.notifications: %+v", req.Notifications)
				log.Warn().Msgf("data: %+v", data)
				firebaseClient.SendNotification(context.Background(), deviceInfo.PushToken, string(deviceInfo.Platform), req.Notifications.CollapseKey, req.Notifications.Title, data["avatar_url"], req.Notifications.Message, req.Notifications.Type, data)
				log.Info().Msg("Push notification sent successfully")
			}

			log.Info().Msg("Successfully created notification")
			return resp, nil
		}, false)); err != nil {
		log.Error().Err(err).Msg("Failed to register RPC command for createNotification")
		return err
	}

	if err := httpRouter.GetRpcServiceEndpoint().RegisterRpcCommand(router.NewServiceCommand(sendCustomEmail,
		func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
			log.Info().Msg("Starting RPC command for Sending Custom Email")

			var req mail.GenericEmailRPC

			if err := json.Unmarshal(request, &req); err != nil {
				return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
			}

			return notification_handler.CreateNotificationResponse{
				Status: true,
			}, nil
		}, false)); err != nil {
		return err
	}

	if err := httpRouter.GetRpcServiceEndpoint().RegisterRpcCommand(router.NewServiceCommand(sendCustomEmailWithGenericHTML,
		func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
			log.Info().Msg("Starting RPC command for Sending Custom Email with Generic HTML")

			var req mail.GenericHTMLEmailRPC

			if err := json.Unmarshal(request, &req); err != nil {
				return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
			}

			return notification_handler.CreateNotificationResponse{
				Status: true,
			}, nil
		}, false)); err != nil {
		return err
	}

	return nil
}
