package api

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/notification-handler/pkg/notification"
)

func InitInternalNotificationApi(httpRouter *router.HttpRouter, apiDef map[string]swagger.ApiDescription) error {
	getNotificationsReadCount := "GetNotificationsReadCount"

	if err := httpRouter.GetRpcAdminEndpoint().RegisterRpcCommand(router.NewAdminCommand(getNotificationsReadCount,
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
		}, common.AccessLevelPublic, "notifications:read_count:view")); err != nil {
		return err
	}

	apiDef[getNotificationsReadCount] = swagger.ApiDescription{
		Request: notification.GetNotificationsReadCountRequest{},
		Tags:    []string{"notification"},
	}

	return nil
}
