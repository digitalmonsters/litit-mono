package api

import (
	"encoding/json"
	"github.com/digitalmonsters/ads-manager/pkg/ad_campaign/ad_moderation"
	commonPkg "github.com/digitalmonsters/ads-manager/pkg/common"
	"github.com/digitalmonsters/ads-manager/pkg/database"
	"github.com/digitalmonsters/go-common/callback"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
)

func (a *apiApp) initAdminApi(httpRouter router.IRpcEndpoint) error {
	restCommands := []router.ICommand{
		a.getModerationRequest(),
		a.setAdsRejectReason(),
	}

	for _, c := range restCommands {
		if err := httpRouter.RegisterRpcCommand(c); err != nil {
			return err
		}
	}

	return nil
}

func (a *apiApp) getModerationRequest() router.ICommand {
	methodName := "GetAdsModerationRequests"

	a.apiDef[methodName] = swagger.ApiDescription{
		Request:  ad_moderation.GetAdModerationRequest{},
		Response: ad_moderation.GetAdModerationResponse{},
		Tags:     []string{"ad moderation"},
	}

	return router.NewAdminCommand(methodName, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req ad_moderation.GetAdModerationRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}
		db := database.GetDbWithContext(database.DbTypeReadonly, executionData.Context)

		resp, err := a.adModerationService.GetAdModerationRequests(req, db, executionData.Context)
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}
		return resp, nil
	}, common.AccessLevelRead, "ads:moderation:view")
}

func (a *apiApp) setAdsRejectReason() router.ICommand {
	methodName := "SetAdsRejectReason"

	a.apiDef[methodName] = swagger.ApiDescription{
		Request:  ad_moderation.SetAdRejectReasonRequest{},
		Response: commonPkg.AddModerationItem{},
		Tags:     []string{"ad moderation"},
	}

	return router.NewAdminCommand(methodName, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req ad_moderation.SetAdRejectReasonRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}
		tx := database.GetDbWithContext(database.DbTypeMaster, executionData.Context).Begin()
		defer tx.Rollback()

		resp, callbacks, err := a.adModerationService.SetAdRejectReason(req, tx, executionData.Context)

		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}
		if err := tx.Commit().Error; err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}
		callback.ExecuteCallbacks(executionData.Context, callbacks...)
		return resp, nil
	}, common.AccessLevelWrite, "ads:moderation:modify")
}
