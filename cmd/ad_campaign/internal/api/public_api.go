package api

import (
	"encoding/json"
	"github.com/digitalmonsters/ads-manager/pkg/ad_campaign"
	"github.com/digitalmonsters/ads-manager/pkg/database"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/pkg/errors"
)

func (a *apiApp) initPublicApi(httpRouter *router.HttpRouter) error {
	restCommands := []*router.RestCommand{
		a.createAdCampaign(),
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

		if req.Budget == 0 {
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
