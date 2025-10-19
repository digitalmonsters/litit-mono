package ad_campaign

import (
	"github.com/digitalmonsters/ads-manager/cmd/ad_campaign/internal/api"
	"github.com/digitalmonsters/ads-manager/pkg/ad_campaign"
	"github.com/digitalmonsters/ads-manager/pkg/ad_campaign/ad_moderation"
	"github.com/digitalmonsters/go-common/application"
	"github.com/digitalmonsters/go-common/router"
)

func Application(
	httpRouter *router.HttpRouter,
	adCampaignService ad_campaign.IService,
	adModerationService ad_moderation.IService,
) *application.BaseApplication {
	return application.NewBaseApplication("ad_campaign").
		AddSubApplication(api.SubApp(httpRouter, adCampaignService, adModerationService))
}
