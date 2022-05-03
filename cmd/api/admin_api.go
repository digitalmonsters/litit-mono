package api

import (
	"encoding/json"
	"github.com/digitalmonsters/configurator/pkg/database"
	"github.com/digitalmonsters/configurator/pkg/feature_toggle"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
)

func InitAdminApi(adminEndpoint router.IRpcEndpoint, apiDesc map[string]swagger.ApiDescription) error {
	if err := adminEndpoint.RegisterRpcCommand(router.NewAdminCommand("GetFeatureToggles", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {

		var req feature_toggle.GetFeatureTogglesRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}
		if resp, err := feature_toggle.GetFeatureToggles(database.GetDbWithContext(database.DbTypeReadonly,
			executionData.Context), req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		} else {
			return resp, nil
		}
	}, common.AccessLevelRead, "feature_toggle:list")); err != nil {
		return err
	}

	if err := adminEndpoint.RegisterRpcCommand(router.NewAdminCommand("GetFeatureToggleEvents", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {

		var req feature_toggle.ListFeatureToggleEventsRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}
		if resp, err := feature_toggle.ListFeatureToggleEvents(database.GetDbWithContext(database.DbTypeReadonly,
			executionData.Context), req.Limit, req.Offset); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		} else {
			return resp, nil
		}
	}, common.AccessLevelRead, "feature_toggle:list")); err != nil {
		return err
	}

	if err := adminEndpoint.RegisterRpcCommand(router.NewAdminCommand("AddFeatureToggle", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {

		var req feature_toggle.CreateFeatureToggleRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}
		if resp, err := feature_toggle.CreateFeatureToggle(database.GetDbWithContext(database.DbTypeReadonly,
			executionData.Context), req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		} else {
			return resp, nil
		}
	}, common.AccessLevelWrite, "feature_toggle:modify")); err != nil {
		return err
	}
	if err := adminEndpoint.RegisterRpcCommand(router.NewAdminCommand("UpdateFeatureToggle", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {

		var req feature_toggle.UpdateFeatureToggleRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}
		if resp, err := feature_toggle.UpdateFeatureToggle(database.GetDbWithContext(database.DbTypeReadonly,
			executionData.Context), req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		} else {
			return resp, nil
		}
	}, common.AccessLevelWrite, "feature_toggle:modify")); err != nil {
		return err
	}

	if err := adminEndpoint.RegisterRpcCommand(router.NewAdminCommand("DeleteFeatureToggle", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {

		var req feature_toggle.DeleteFeatureToggleRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}
		if err := feature_toggle.DeleteFeatureToggle(database.GetDbWithContext(database.DbTypeReadonly,
			executionData.Context), req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		} else {
			return map[string]interface{}{
				"success": true,
			}, nil
		}
	}, common.AccessLevelRead, "feature_toggle:modify")); err != nil {
		return err
	}

	apiDesc["GetFeatureToggles"] = swagger.ApiDescription{
		Request:           feature_toggle.GetFeatureTogglesRequest{},
		Response:          feature_toggle.GetFeatureTogglesResponse{},
		MethodDescription: "method for getting feature toggles",
		Tags:              []string{"feature toggle"},
	}
	apiDesc["GetFeatureToggleEvents"] = swagger.ApiDescription{
		Request:           feature_toggle.ListFeatureToggleEventsRequest{},
		Response:          feature_toggle.ListFeatureToggleEventsResponse{},
		MethodDescription: "method for getting feature toggle events",
		Tags:              []string{"feature toggle"},
	}
	apiDesc["AddFeatureToggle"] = swagger.ApiDescription{
		Request:           feature_toggle.CreateFeatureToggleRequest{},
		Response:          feature_toggle.FeatureToggleModel{},
		MethodDescription: "method for adding feature toggles configuration",
		Tags:              []string{"feature toggle"},
	}
	apiDesc["UpdateFeatureToggle"] = swagger.ApiDescription{
		Request:           feature_toggle.UpdateFeatureToggleRequest{},
		Response:          feature_toggle.FeatureToggleModel{},
		MethodDescription: "method for updating feature toggles configuration",
		Tags:              []string{"feature toggle"},
	}
	apiDesc["DeleteFeatureToggles"] = swagger.ApiDescription{
		Request:           feature_toggle.DeleteFeatureToggleRequest{},
		Response:          nil,
		MethodDescription: "method for deleting feature toggles",
		Tags:              []string{"feature toggle"},
	}

	return nil
}
