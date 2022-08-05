package api

import (
	"encoding/json"
	"github.com/digitalmonsters/ads-manager/pkg/database"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/digitalmonsters/go-common/wrappers/ads_manager"
)

func (a *apiApp) initInternalApi(httpRouter router.IRpcEndpoint) error {
	restCommands := []router.ICommand{
		a.getAdsContentForUser(),
	}

	for _, c := range restCommands {
		if err := httpRouter.RegisterRpcCommand(c); err != nil {
			return err
		}
	}

	return nil
}

func (a *apiApp) getAdsContentForUser() router.ICommand {
	methodName := "GetAdsContentForUser"

	a.apiDef[methodName] = swagger.ApiDescription{
		Request:  ads_manager.GetAdsContentForUserRequest{},
		Response: ads_manager.GetAdsContentForUserResponse{},
		Tags:     []string{"ads_content"},
	}

	return router.NewServiceCommand(methodName,
		func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
			var req ads_manager.GetAdsContentForUserRequest

			if err := json.Unmarshal(request, &req); err != nil {
				return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
			}

			resp, err := a.adCampaignService.GetAdsContentForUser(req,
				database.GetDbWithContext(database.DbTypeReadonly, executionData.Context), executionData.Context)
			if err != nil {
				return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
			}

			return resp, nil
		}, false)
}
