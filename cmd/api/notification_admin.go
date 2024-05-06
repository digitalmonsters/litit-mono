package api

import (
	"encoding/json"
	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/go-common/wrappers/follow"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/notification-handler/configs"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/notification"
)

func InitAdminNotificationApi(httpRouter *router.HttpRouter, apiDef map[string]swagger.ApiDescription,
	userGoWrapper user_go.IUserGoWrapper, followWrapper follow.IFollowWrapper, jobber *machinery.Server) error {
	listNotificationsMethod := "ListNotifications"
	generalPushNotificationTaskMethod := "GeneralPushNotificationTask"
	userPushNotificationTaskMethod := "UserPushNotificationTask"

	if err := httpRouter.GetRpcAdminEndpoint().RegisterRpcCommand(router.NewAdminCommand(listNotificationsMethod,
		func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
			var req notification.ListNotificationsByAdminRequest

			if err := json.Unmarshal(request, &req); err != nil {
				return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
			}

			resp, err := notification.ListNotificationsByAdmin(
				database.GetDbWithContext(database.DbTypeReadonly, executionData.Context), req, userGoWrapper,
				followWrapper, executionData.Context)
			if err != nil {
				return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
			}

			return resp, nil
		}, common.AccessLevelPublic, "notifications:list")); err != nil {
		return err
	}

	if err := httpRouter.GetRpcAdminEndpoint().RegisterRpcCommand(router.NewAdminCommand(generalPushNotificationTaskMethod,
		func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
			var req GeneralPushNotificationTaskRequest

			if err := json.Unmarshal(request, &req); err != nil {
				return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
			}

			if _, err := jobber.SendTask(&tasks.Signature{
				Name: string(configs.GeneralPushNotificationTask),
				Args: []tasks.Arg{
					{
						Name:  "currentDate",
						Type:  "string",
						Value: req.CurrentDate,
					},
				}}); err != nil {
				return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
			}

			return nil, nil
		}, common.AccessLevelWrite, "notifications:task:general")); err != nil {
		return err
	}

	if err := httpRouter.GetRpcAdminEndpoint().RegisterRpcCommand(router.NewAdminCommand(userPushNotificationTaskMethod,
		func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
			var req UserPushNotificationTaskRequest

			if err := json.Unmarshal(request, &req); err != nil {
				return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
			}

			itemMarshalled, err := json.Marshal(req.Item)
			if err != nil {
				return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
			}

			if _, err = jobber.SendTask(&tasks.Signature{
				Name: string(configs.UserPushNotificationTask),
				Args: []tasks.Arg{
					{
						Name:  "currentDate",
						Type:  "string",
						Value: req.CurrentDate,
					},
					{
						Name:  "item",
						Type:  "string",
						Value: string(itemMarshalled),
					},
					{
						Name:  "traceHeader",
						Type:  "string",
						Value: "",
					},
				}}); err != nil {
				return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
			}

			return nil, nil
		}, common.AccessLevelWrite, "notifications:task:user")); err != nil {
		return err
	}

	apiDef[listNotificationsMethod] = swagger.ApiDescription{
		Request:  notification.ListNotificationsByAdminRequest{},
		Response: notification.ListNotificationsByAdminResponse{},
		Tags:     []string{"notification"},
	}

	apiDef[generalPushNotificationTaskMethod] = swagger.ApiDescription{
		Request: GeneralPushNotificationTaskRequest{},
		Tags:    []string{"notification"},
	}

	apiDef[userPushNotificationTaskMethod] = swagger.ApiDescription{
		Request: UserPushNotificationTaskRequest{},
		Tags:    []string{"notification"},
	}

	return nil
}
