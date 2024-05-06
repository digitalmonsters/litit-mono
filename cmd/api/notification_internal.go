package api

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/go-common/wrappers/notification_handler"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/notification"
)

func InitInternalNotificationApi(httpRouter *router.HttpRouter, apiDef map[string]swagger.ApiDescription) error {
	getNotificationsReadCount := "GetNotificationsReadCount"
	disableUnregisteredTokens := "DisableUnregisteredTokens"

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

	return nil
}
