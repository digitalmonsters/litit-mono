package indecent_content_checker

import (
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/digitalmonsters/go-common/wrappers"
	"github.com/rs/zerolog/log"
	"go.elastic.co/apm"
	"net/url"
	"time"
)

func NewIndecentContentCheckerWrapper(config boilerplate.WrapperConfig) IIndecentContentCheckerWrapper {
	timeout := 5 * time.Second

	if config.TimeoutSec > 0 {
		timeout = time.Duration(config.TimeoutSec) * time.Second
	}

	if len(config.ApiUrl) == 0 {
		config.ApiUrl = "http://indecent_content_checker"

		log.Warn().Msgf("Api Url is missing for IndecentContentChecker. Setting as default : %v", config.ApiUrl)
	}

	return &IndecentContentCheckerWrapper{
		baseWrapper:    wrappers.GetBaseWrapper(),
		defaultTimeout: timeout,
		apiUrl:         common.StripSlashFromUrl(config.ApiUrl),
		serviceName:    "indecent_content_checker",
	}
}

func (w *IndecentContentCheckerWrapper) GetPredictions(req GetPredictionsRequest, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[[]PredictionItem] {
	resChan := make(chan wrappers.GenericResponseChan[[]PredictionItem], 2)

	go func() {
		finalResponse := wrappers.GenericResponseChan[[]PredictionItem]{}
		var predictions []PredictionItem

		params := url.Values{}
		params.Add("url", req.ImageUrl)
		requestUrl := fmt.Sprintf("%v/predictions?%v", w.apiUrl, params.Encode())

		rpcInternalResponse := <-w.baseWrapper.SendRequestWithRpcResponseFromNodeJsService(
			requestUrl, "GET", "application/json", "get predictions",
			nil, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog,
		)

		finalResponse.Error = rpcInternalResponse.Error

		if finalResponse.Error == nil && len(rpcInternalResponse.Result) > 0 {
			if err := json.Unmarshal(rpcInternalResponse.Result, &predictions); err != nil {
				finalResponse.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    w.baseWrapper.GetHostName(),
					ServiceName: w.serviceName,
				}
			} else {
				finalResponse.Response = predictions
			}
		}
		resChan <- finalResponse
	}()

	return resChan
}
