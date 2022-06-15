package api

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/go-common/wrappers/follow"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/digitalmonsters/notification-handler/pkg/notification"
)

func InitAdminNotificationApi(httpRouter *router.HttpRouter, apiDef map[string]swagger.ApiDescription,
	userGoWrapper user_go.IUserGoWrapper, followWrapper follow.IFollowWrapper) error {
	listNotificationsMethod := "ListNotifications"

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

	apiDef[listNotificationsMethod] = swagger.ApiDescription{
		Request:  notification.ListNotificationsByAdminRequest{},
		Response: notification.ListNotificationsByAdminResponse{},
		Tags:     []string{"notification"},
	}

	return nil
}
