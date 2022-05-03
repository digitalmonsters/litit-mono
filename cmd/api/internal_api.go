package api

import (
	"encoding/json"
	"github.com/digitalmonsters/configurator/pkg/database"
	"github.com/digitalmonsters/configurator/pkg/feature_toggle"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
)

func InitInternalApi(serviceEndpoint router.IRpcEndpoint, apiDesc map[string]swagger.ApiDescription) error {
	if err := serviceEndpoint.RegisterRpcCommand(router.NewServiceCommand("InternalGetFeatureToggles", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		if resp, err := feature_toggle.GetAllFeatureToggles(database.GetDbWithContext(database.DbTypeReadonly,
			executionData.Context)); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		} else {
			return resp, nil
		}
	}, false)); err != nil {
		return err
	}

	if err := serviceEndpoint.RegisterRpcCommand(router.NewServiceCommand("InternalCreateFeatureToggleEvent", func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req feature_toggle.CreateFeatureToggleEventsRequest
		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}
		if err := feature_toggle.CreateFeatureToggleEvents(database.GetDbWithContext(database.DbTypeMaster,
			executionData.Context), req.Events); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		} else {
			return map[string]interface{}{
				"success": true,
			}, nil
		}
	}, false)); err != nil {
		return err
	}

	apiDesc["InternalGetFeatureToggles"] = swagger.ApiDescription{
		Response:          map[string]database.FeatureToggleConfig{},
		MethodDescription: "internal method for getting all feature toggles",
		Tags:              []string{"internal"},
	}
	apiDesc["InternalCreateFeatureToggleEvent"] = swagger.ApiDescription{
		Request:           feature_toggle.CreateFeatureToggleEventsRequest{},
		MethodDescription: "internal method for creating feature toggle event",
		Tags:              []string{"internal"},
	}

	return nil
}
