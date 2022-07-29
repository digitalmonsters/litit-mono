package api

import (
	"github.com/digitalmonsters/ads-manager/pkg/ad_campaign"
	"github.com/digitalmonsters/ads-manager/pkg/ad_campaign/ad_moderation"
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/swagger"
	"github.com/rs/zerolog"
)

type apiApp struct {
	httpRouter          *router.HttpRouter
	apiDef              map[string]swagger.ApiDescription
	adCampaignService   ad_campaign.IService
	adModerationService ad_moderation.IService
}

func SubApp(
	httpRouter *router.HttpRouter,
	apiDef map[string]swagger.ApiDescription,
	adCampaignService ad_campaign.IService,
	adModerationService ad_moderation.IService,
) application.SubApplication {
	return &apiApp{
		httpRouter:          httpRouter,
		apiDef:              apiDef,
		adCampaignService:   adCampaignService,
		adModerationService: adModerationService,
	}
}

func (a *apiApp) Init(subAppLogger zerolog.Logger) error {
	if err := a.initPublicApi(a.httpRouter); err != nil {
		return err
	}

	if err := a.initAdminApi(a.httpRouter.GetRpcAdminEndpoint()); err != nil {
		return err
	}

	return nil
}

func (a *apiApp) Name() string {
	return "api"
}

func (a *apiApp) Close() error {
	return nil
}
