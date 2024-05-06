package settings

import (
	"encoding/json"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	settingsPkg "github.com/digitalmonsters/notification-handler/pkg/settings"
)

func (a settingsApp) initAdminApi(endpoint router.IRpcEndpoint) error {
	commands := []router.ICommand{
		a.getPushSettingsByAdmin(),
		a.changePushSettingsByAdmin(),
	}

	for _, c := range commands {
		if err := endpoint.RegisterRpcCommand(c); err != nil {
			return err
		}
	}

	return nil
}

func (a settingsApp) getPushSettingsByAdmin() router.ICommand {
	method := "GetPushSettings"

	a.apiDef[method] = swagger.ApiDescription{
		Request:           settingsPkg.GetPushSettingsByAdminRequest{},
		Response:          map[string]settingsPkg.GetPushSettingsByAdminItem{},
		MethodDescription: "get push notifications settings",
		Tags:              []string{"settings"},
	}

	return router.NewAdminCommand(method, func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req settingsPkg.GetPushSettingsByAdminRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		resp, err := a.settingsService.GetPushSettingsByAdmin(req, executionData.Context, database.GetDbWithContext(database.DbTypeReadonly, executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return resp, nil
	}, common.AccessLevelRead, "notification:user:settings")
}

func (a settingsApp) changePushSettingsByAdmin() router.ICommand {
	method := "ChangePushSettings"

	a.apiDef[method] = swagger.ApiDescription{
		Request:           settingsPkg.ChangePushSettingsByAdminRequest{},
		MethodDescription: "change push notifications settings",
		Tags:              []string{"settings"},
	}

	return router.NewAdminCommand(method, func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req settingsPkg.ChangePushSettingsByAdminRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		if err := a.settingsService.ChangePushSettings(req.Settings, req.UserId, executionData.Context); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return nil, nil
	}, common.AccessLevelWrite, "notification:user:settings")
}
