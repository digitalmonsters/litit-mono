package api

import (
	"encoding/json"
	"github.com/digitalmonsters/configurator/pkg/configs"
	"github.com/digitalmonsters/configurator/pkg/database"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
)

func (a *apiApp) initAdminApi(endpoint router.IRpcEndpoint) error {
	commands := []router.ICommand{
		a.getConfigs(),
		a.getConfigLogs(),
		a.upsertConfig(),
	}

	for _, c := range commands {
		if err := endpoint.RegisterRpcCommand(c); err != nil {
			return err
		}
	}

	return nil
}

func (a *apiApp) getConfigs() router.ICommand {

	a.apiDef["GetConfigs"] = swagger.ApiDescription{
		Request:           configs.GetConfigRequest{},
		Response:          configs.GetConfigResponse{},
		MethodDescription: "Get configs on admin panel",
		Summary:           "Get configs",
		Tags:              []string{"configs", "admin"},
	}

	return router.NewAdminCommand("GetConfigs", func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req configs.GetConfigRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}
		resp, err := a.service.AdminGetConfigs(database.GetDb(database.DbTypeReadonly).WithContext(executionData.Context), req)
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}
		return resp, nil
	}, common.AccessLevelRead, "configs:view")
}

func (a *apiApp) getConfigLogs() router.ICommand {

	a.apiDef["GetConfigLogs"] = swagger.ApiDescription{
		Request:           configs.GetConfigLogsRequest{},
		Response:          configs.GetConfigLogsResponse{},
		MethodDescription: "Get config logs on admin panel",
		Summary:           "Get config logs",
		Tags:              []string{"configs", "logs", "admin"},
	}

	return router.NewAdminCommand("GetConfigLogs", func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req configs.GetConfigLogsRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}
		resp, err := a.service.AdminGetConfigLogs(database.GetDb(database.DbTypeReadonly).WithContext(executionData.Context), req)
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}
		return resp, nil
	}, common.AccessLevelRead, "config:logs:view")
}

func (a *apiApp) upsertConfig() router.ICommand {

	a.apiDef["UpsertConfig"] = swagger.ApiDescription{
		Request:           configs.UpsertConfigRequest{},
		Response:          configs.ConfigModel{},
		MethodDescription: "Upsert config on admin panel",
		Summary:           "Upsert config",
		Tags:              []string{"configs", "admin"},
	}

	return router.NewAdminCommand("UpsertConfig", func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req configs.UpsertConfigRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}
		resp, err := a.service.AdminUpsertConfig(database.GetDb(database.DbTypeMaster).WithContext(executionData.Context),
			req, executionData.UserId, a.publisher, executionData.Context)
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}
		return resp, nil
	}, common.AccessLevelWrite, "configs:upsert")
}
