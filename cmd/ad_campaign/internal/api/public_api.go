package api

import (
	"context"
	"encoding/json"
	"github.com/digitalmonsters/ads-manager/pkg/ad_campaign"
	"github.com/digitalmonsters/ads-manager/pkg/database"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

func (a *apiApp) initPublicApi(httpRouter *router.HttpRouter) error {
	restCommands := []*router.RestCommand{
		a.createAdCampaign(),
		a.clickLink(),
		a.stopAdCampaign(),
		a.startAdCampaign(),
		a.listAdCampaigns(),
		a.hasAdCampaigns(),
	}

	for _, c := range restCommands {
		if err := httpRouter.RegisterRestCmd(c); err != nil {
			return err
		}
	}

	return nil
}

func (a *apiApp) createAdCampaign() *router.RestCommand {
	path := "/create_ad_campaign"

	a.apiDef[path] = swagger.ApiDescription{
		Request: ad_campaign.CreateAdCampaignRequest{},
		Tags:    []string{"ad_campaign"},
	}

	return router.NewRestCommand(func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req ad_campaign.CreateAdCampaignRequest

		if len(request) > 0 {
			if err := json.Unmarshal(request, &req); err != nil {
				return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
			}
		}

		if req.AdType < 1 || req.AdType > 2 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid ad_type"), error_codes.GenericValidationError)
		}

		if req.ContentId == 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid content_id"), error_codes.GenericValidationError)
		}

		if req.Budget.LessThanOrEqual(decimal.Zero) {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid budget"), error_codes.GenericValidationError)
		}

		tx := database.GetDbWithContext(database.DbTypeMaster, executionData.Context).Begin()
		defer tx.Rollback()

		if err := a.adCampaignService.CreateAdCampaign(req, executionData.UserId, tx, executionData.Context); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		if err := tx.Commit().Error; err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return nil, nil
	}, path, router.MethodPost).RequireIdentityValidation().Build()
}

func (a *apiApp) clickLink() *router.RestCommand {
	path := "/click_link"

	a.apiDef[path] = swagger.ApiDescription{
		Request: ad_campaign.ClickLinkRequest{},
		Tags:    []string{"ad_campaign"},
	}

	return router.NewRestCommand(func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req ad_campaign.ClickLinkRequest

		if len(request) > 0 {
			if err := json.Unmarshal(request, &req); err != nil {
				return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
			}
		}

		if req.ContentId == 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid content_id"), error_codes.GenericValidationError)
		}

		tx := database.GetDbWithContext(database.DbTypeMaster, executionData.Context).Begin()
		defer tx.Rollback()

		if err := a.adCampaignService.ClickLink(executionData.UserId, req, tx); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		if err := tx.Commit().Error; err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return nil, nil
	}, path, router.MethodPost).RequireIdentityValidation().Build()
}

func (a *apiApp) stopAdCampaign() *router.RestCommand {
	path := "/stop_ad_campaign"

	a.apiDef[path] = swagger.ApiDescription{
		Request: ad_campaign.StopAdCampaignRequest{},
		Tags:    []string{"ad_campaign"},
	}

	return router.NewRestCommand(func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req ad_campaign.StopAdCampaignRequest

		if len(request) > 0 {
			if err := json.Unmarshal(request, &req); err != nil {
				return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
			}
		}

		if req.AdCampaignId == 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid ad_campaign_id"), error_codes.GenericValidationError)
		}

		tx := database.GetDbWithContext(database.DbTypeMaster, executionData.Context).Begin()
		defer tx.Rollback()

		if err := a.adCampaignService.StopAdCampaign(executionData.UserId, req, tx); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		if err := tx.Commit().Error; err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return nil, nil
	}, path, router.MethodPost).RequireIdentityValidation().Build()
}

func (a *apiApp) startAdCampaign() *router.RestCommand {
	path := "/start_ad_campaign"

	a.apiDef[path] = swagger.ApiDescription{
		Request: ad_campaign.StartAdCampaignRequest{},
		Tags:    []string{"ad_campaign"},
	}

	return router.NewRestCommand(func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req ad_campaign.StartAdCampaignRequest

		if len(request) > 0 {
			if err := json.Unmarshal(request, &req); err != nil {
				return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
			}
		}

		if req.AdCampaignId == 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid ad_campaign_id"), error_codes.GenericValidationError)
		}

		tx := database.GetDbWithContext(database.DbTypeMaster, executionData.Context).Begin()
		defer tx.Rollback()

		if err := a.adCampaignService.StartAdCampaign(executionData.UserId, req, tx, context.TODO()); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		if err := tx.Commit().Error; err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return nil, nil
	}, path, router.MethodPost).RequireIdentityValidation().Build()
}

func (a *apiApp) listAdCampaigns() *router.RestCommand {
	path := "/list_ad_campaigns"

	a.apiDef[path] = swagger.ApiDescription{
		Request:  ad_campaign.ListAdCampaignsRequest{},
		Response: ad_campaign.ListAdCampaignsResponse{},
		Tags:     []string{"ad_campaign"},
	}

	return router.NewRestCommand(func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		var req ad_campaign.ListAdCampaignsRequest

		if len(request) > 0 {
			if err := json.Unmarshal(request, &req); err != nil {
				return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
			}
		}

		resp, err := a.adCampaignService.ListAdCampaigns(executionData.UserId, req, database.GetDbWithContext(database.DbTypeReadonly, executionData.Context), executionData.Context)
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return resp, nil
	}, path, router.MethodPost).RequireIdentityValidation().Build()
}

func (a *apiApp) hasAdCampaigns() *router.RestCommand {
	path := "/has_ad_campaigns"

	a.apiDef[path] = swagger.ApiDescription{
		Response: ad_campaign.HasAdCampaignsResponse{},
		Tags:     []string{"ad_campaign"},
	}

	return router.NewRestCommand(func(request []byte, executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		resp, err := a.adCampaignService.HasAdCampaigns(executionData.UserId, database.GetDbWithContext(database.DbTypeReadonly, executionData.Context))
		if err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		}

		return resp, nil
	}, path, router.MethodGet).RequireIdentityValidation().Build()
}
