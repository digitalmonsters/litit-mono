package settings

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"net/http"
)

func (a settingsApp) initPublicApi(httpRouter *router.HttpRouter) error {
	restCommands := []*router.RestCommand{
		a.getPushSettings(),
		a.changePushSettings(),
	}

	for _, c := range restCommands {
		if err := httpRouter.RegisterRestCmd(c); err != nil {
			return err
		}
	}

	return nil
}

func (a settingsApp) getPushSettings() *router.RestCommand {
	path := "/mobile/v1/push_settings"

	a.apiDef[path] = swagger.ApiDescription{
		Response:          map[string]bool{},
		MethodDescription: "get push notifications settings",
		Tags:              []string{"settings"},
	}

	return router.NewRestCommand(func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		resp, err := a.settingsService.GetPushSettings(executionData.UserId, executionData.Context,
			database.GetDbWithContext(database.DbTypeReadonly, executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		return resp, nil
	}, path, http.MethodGet).RequireIdentityValidation().Build()
}

func (a settingsApp) changePushSettings() *router.RestCommand {
	path := "/mobile/v1/change_push_settings"

	a.apiDef[path] = swagger.ApiDescription{
		Request:           map[string]bool{},
		MethodDescription: "change push notifications settings",
		Tags:              []string{"settings"},
	}

	return router.NewRestCommand(func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req map[string]bool

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}
		if err := a.settingsService.ChangePushSettings(req, executionData.UserId, executionData.Context); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		return nil, nil
	}, path, http.MethodPost).RequireIdentityValidation().Build()
}
