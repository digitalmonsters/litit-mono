package bot_factory

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"go.elastic.co/apm"

	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/wrappers"
)

type IBotFactory interface {
	SetSuperInfluencer(userId int64, ctx context.Context,
		forceLog bool) chan wrappers.GenericResponseChan[SetSuperInfluencerResponse]
}

//goland:noinspection GoNameStartsWithPackageName
type BotFactoryWrapper struct {
	baseWrapper    *wrappers.BaseWrapper
	defaultTimeout time.Duration
	apiUrl         string
	serviceName    string
}

func NewAdsManagerWrapper(config boilerplate.WrapperConfig) IBotFactory {
	timeout := 5 * time.Second

	if config.TimeoutSec > 0 {
		timeout = time.Duration(config.TimeoutSec) * time.Second
	}

	if len(config.ApiUrl) == 0 {
		config.ApiUrl = "http://bot-factory"

		log.Warn().Msgf("Api Url is missing for Ads-Manager. Setting as default : %v", config.ApiUrl)
	}

	return &BotFactoryWrapper{
		baseWrapper:    wrappers.GetBaseWrapper(),
		defaultTimeout: timeout,
		apiUrl:         fmt.Sprintf("%v/rpc-service", common.StripSlashFromUrl(config.ApiUrl)),
		serviceName:    "bot_factory",
	}
}

func (w *BotFactoryWrapper) SetSuperInfluencer(userId int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[SetSuperInfluencerResponse] {
	return wrappers.ExecuteRpcRequestAsync[SetSuperInfluencerResponse](w.baseWrapper, w.apiUrl, "GetAdsContentForUser",
		SetSuperInfluencerRequest{
			UserId: userId,
		}, map[string]string{}, w.defaultTimeout, apm.TransactionFromContext(ctx), w.serviceName, forceLog)
}
