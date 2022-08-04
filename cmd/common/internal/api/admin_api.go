package api

import (
	"encoding/json"
	commonPkg "github.com/digitalmonsters/ads-manager/pkg/common"
	"github.com/digitalmonsters/ads-manager/pkg/database"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
)

func (a *apiApp) initAdminApi(httpRouter router.IRpcEndpoint) error {
	restCommands := []router.ICommand{
		a.upsertActionButtons(),
		a.upsertRejectReasons(),
		a.deleteActionButtons(),
		a.deleteRejectReasons(),
		a.listRejectReasons(),
		a.listActionButtons(),
	}

	for _, c := range restCommands {
		if err := httpRouter.RegisterRpcCommand(c); err != nil {
			return err
		}
	}

	return nil
}

func (a *apiApp) upsertActionButtons() router.ICommand {
	methodName := "UpsertActionButtons"

	a.apiDef[methodName] = swagger.ApiDescription{
		Request: commonPkg.UpsertActionButtonsRequest{},
		Tags:    []string{"common", "button"},
	}

	return router.NewAdminCommand(methodName, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req commonPkg.UpsertActionButtonsRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		tx := database.GetDbWithContext(database.DbTypeMaster, executionData.Context).Begin()
		defer tx.Rollback()

		if err := a.commonService.UpsertActionButtons(req, tx); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		if err := tx.Commit().Error; err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return nil, nil
	}, common.AccessLevelWrite, "ads:action_button:modify")
}

func (a *apiApp) upsertRejectReasons() router.ICommand {
	methodName := "UpsertRejectReasons"

	a.apiDef[methodName] = swagger.ApiDescription{
		Request: commonPkg.UpsertRejectReasonsRequest{},
		Tags:    []string{"common", "reject reason"},
	}

	return router.NewAdminCommand(methodName, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req commonPkg.UpsertRejectReasonsRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		tx := database.GetDbWithContext(database.DbTypeMaster, executionData.Context).Begin()
		defer tx.Rollback()

		if err := a.commonService.UpsertRejectReasons(req, tx); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		if err := tx.Commit().Error; err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return nil, nil
	}, common.AccessLevelWrite, "ads:reject_reason:modify")
}

func (a *apiApp) deleteActionButtons() router.ICommand {
	methodName := "DeleteActionButtons"

	a.apiDef[methodName] = swagger.ApiDescription{
		Request: commonPkg.DeleteRequest{},
		Tags:    []string{"common", "button"},
	}

	return router.NewAdminCommand(methodName, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req commonPkg.DeleteRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		tx := database.GetDbWithContext(database.DbTypeMaster, executionData.Context).Begin()
		defer tx.Rollback()

		if err := a.commonService.DeleteActionButtons(req, tx); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		if err := tx.Commit().Error; err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return nil, nil
	}, common.AccessLevelWrite, "ads:action_button:modify")
}

func (a *apiApp) deleteRejectReasons() router.ICommand {
	methodName := "DeleteRejectReasons"

	a.apiDef[methodName] = swagger.ApiDescription{
		Request: commonPkg.DeleteRequest{},
		Tags:    []string{"common", "reject reason"},
	}

	return router.NewAdminCommand(methodName, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req commonPkg.DeleteRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		tx := database.GetDbWithContext(database.DbTypeMaster, executionData.Context).Begin()
		defer tx.Rollback()

		if err := a.commonService.DeleteRejectReasons(req, tx); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		if err := tx.Commit().Error; err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return nil, nil
	}, common.AccessLevelWrite, "ads:reject_reason:modify")
}

func (a *apiApp) listActionButtons() router.ICommand {
	methodName := "ListActionButtons"

	a.apiDef[methodName] = swagger.ApiDescription{
		Request:  commonPkg.ListActionButtonsRequest{},
		Response: commonPkg.ListActionButtonsResponse{},
		Tags:     []string{"common", "button"},
	}

	return router.NewAdminCommand(methodName, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req commonPkg.ListActionButtonsRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		db := database.GetDbWithContext(database.DbTypeReadonly, executionData.Context)

		resp, err := a.commonService.ListActionButtons(req, db)
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return resp, nil
	}, common.AccessLevelRead, "ads:action_button:view")
}

func (a *apiApp) listRejectReasons() router.ICommand {
	methodName := "ListRejectReasons"

	a.apiDef[methodName] = swagger.ApiDescription{
		Request:  commonPkg.ListRejectReasonsRequest{},
		Response: commonPkg.ListRejectReasonsResponse{},
		Tags:     []string{"common", "reject reason"},
	}

	return router.NewAdminCommand(methodName, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req commonPkg.ListRejectReasonsRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		db := database.GetDbWithContext(database.DbTypeReadonly, executionData.Context)

		resp, err := a.commonService.ListRejectReasons(req, db)
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}
		return resp, nil
	}, common.AccessLevelRead, "ads:reject_reason:view")
}

func (a *apiApp) upsertAdCampaignCountryPrices() router.ICommand {
	methodName := "UpsertAdCampaignCountryPrices"

	a.apiDef[methodName] = swagger.ApiDescription{
		Request:  commonPkg.UpsertAdCampaignCountryPriceRequest{},
		Response: nil,
		Tags:     []string{"common", "country price"},
	}

	return router.NewAdminCommand(methodName, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req commonPkg.UpsertAdCampaignCountryPriceRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		tx := database.GetDbWithContext(database.DbTypeMaster, executionData.Context).Begin()
		defer tx.Rollback()

		if err := a.commonService.UpsertAdCampaignCountryPrice(req, tx); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		if err := tx.Commit().Error; err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}
		return nil, nil

	}, common.AccessLevelWrite, "ads:country_price:modify")
}

func (a *apiApp) listAdCampaignCountryPrices() router.ICommand {
	methodName := "ListAdCampaignCountryPrices"

	a.apiDef[methodName] = swagger.ApiDescription{
		Request:  commonPkg.ListAdCampaignCountryPriceRequest{},
		Response: commonPkg.ListAdCampaignCountryPriceResponse{},
		Tags:     []string{"common", "country price"},
	}

	return router.NewAdminCommand(methodName, func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req commonPkg.ListAdCampaignCountryPriceRequest

		if err := json.Unmarshal(request, &req); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		db := database.GetDbWithContext(database.DbTypeReadonly, executionData.Context)

		resp, err := a.commonService.ListAdCampaignCountryPrices(req, db)
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}
		return resp, nil
	}, common.AccessLevelRead, "ads:country_price:view")
}
