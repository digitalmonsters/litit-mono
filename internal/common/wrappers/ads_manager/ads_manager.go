package ads_manager

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/wrappers"
	"github.com/rs/zerolog/log"
	"go.elastic.co/apm"
	"time"
)

type IAdsManagerWrapper interface {
	GetAdsContentForUser(userId int64, contentIdsToMix []int64, contentIdsToIgnore []int64, ctx context.Context,
		forceLog bool) chan wrappers.GenericResponseChan[GetAdsContentForUserResponse]
}

//goland:noinspection GoNameStartsWithPackageName
type AdsManagerWrapper struct {
	baseWrapper    *wrappers.BaseWrapper
	defaultTimeout time.Duration
	apiUrl         string
	serviceName    string
}

func NewAdsManagerWrapper(config boilerplate.WrapperConfig) IAdsManagerWrapper {
	timeout := 5 * time.Second

	if config.TimeoutSec > 0 {
		timeout = time.Duration(config.TimeoutSec) * time.Second
	}

	if len(config.ApiUrl) == 0 {
		config.ApiUrl = "http://ads-manager"

		log.Warn().Msgf("Api Url is missing for Ads-Manager. Setting as default : %v", config.ApiUrl)
	}

	return &AdsManagerWrapper{
		baseWrapper:    wrappers.GetBaseWrapper(),
		defaultTimeout: timeout,
		apiUrl:         fmt.Sprintf("%v/rpc-service", common.StripSlashFromUrl(config.ApiUrl)),
		serviceName:    "ads_manager",
	}
}

func (w *AdsManagerWrapper) GetAdsContentForUser(userId int64, contentIdsToMix []int64, contentIdsToIgnore []int64,
	ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[GetAdsContentForUserResponse] {
	return wrappers.ExecuteRpcRequestAsync[GetAdsContentForUserResponse](w.baseWrapper, w.apiUrl, "GetAdsContentForUser",
		GetAdsContentForUserRequest{
			UserId:             userId,
			ContentIdsToMix:    contentIdsToMix,
			ContentIdsToIgnore: contentIdsToIgnore,
		}, map[string]string{}, w.defaultTimeout, apm.TransactionFromContext(ctx), w.serviceName, forceLog)
}
