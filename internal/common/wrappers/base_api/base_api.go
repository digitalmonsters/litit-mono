package base_api

import (
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/digitalmonsters/go-common/wrappers"
	"github.com/patrickmn/go-cache"
	"github.com/rs/zerolog/log"
	"go.elastic.co/apm"
	"time"
)

type IBaseApiWrapper interface {
	GetCountriesWithAgeLimit(apmTransaction *apm.Transaction, forceLog bool) chan GetCountriesWithAgeLimitResponseChan
}

//goland:noinspection GoNameStartsWithPackageName
type BaseApiWrapper struct {
	baseWrapper    *wrappers.BaseWrapper
	defaultTimeout time.Duration
	apiUrl         string
	serviceName    string
	cache          *cache.Cache
}

func NewBaseApiWrapper(config boilerplate.WrapperConfig) IBaseApiWrapper {
	timeout := 5 * time.Second

	if config.TimeoutSec > 0 {
		timeout = time.Duration(config.TimeoutSec) * time.Second
	}

	if len(config.ApiUrl) == 0 {
		config.ApiUrl = "http://base-api"

		log.Warn().Msgf("Api Url is missing for BaseApi. Setting as default : %v", config.ApiUrl)
	}

	return &BaseApiWrapper{
		baseWrapper:    wrappers.GetBaseWrapper(),
		defaultTimeout: timeout,
		apiUrl:         common.StripSlashFromUrl(config.ApiUrl),
		serviceName:    "base-api",
		cache:          cache.New(4*time.Minute, 5*time.Minute),
	}
}

var countriesCacheKey = "countries_with_age_limit"

func (w *BaseApiWrapper) GetCountriesWithAgeLimit(apmTransaction *apm.Transaction,
	forceLog bool) chan GetCountriesWithAgeLimitResponseChan {
	resChan := make(chan GetCountriesWithAgeLimitResponseChan, 2)

	go func() {
		var cachedItems []Country
		items, hasCachedItems := w.cache.Get(countriesCacheKey)
		finalResponse := GetCountriesWithAgeLimitResponseChan{}
		var countries []Country

		if hasCachedItems {
			countries = items.([]Country)

			if len(countries) > 0 {
				finalResponse.Items = cachedItems
				resChan <- finalResponse
				return
			}
		}

		rpcInternalResponse := <-w.baseWrapper.SendRequestWithRpcResponseFromNodeJsService(fmt.Sprintf("%v/mobile/v1/location/getCountriesWithAgeLimit", w.apiUrl),
			"GET",
			"application/json",
			"get countries",
			nil, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

		finalResponse.Error = rpcInternalResponse.Error

		if finalResponse.Error == nil && len(rpcInternalResponse.Result) > 0 {
			if err := json.Unmarshal(rpcInternalResponse.Result, &countries); err != nil {
				finalResponse.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    w.baseWrapper.GetHostName(),
					ServiceName: w.serviceName,
				}
			} else {
				w.cache.Set(countriesCacheKey, countries, cache.DefaultExpiration)

				finalResponse.Items = countries
			}
		}
		resChan <- finalResponse
	}()

	return resChan
}
