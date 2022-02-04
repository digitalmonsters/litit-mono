package go_tokenomics

import (
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/wrappers"
	"github.com/rs/zerolog/log"
	"go.elastic.co/apm"
	"time"
)

type Wrapper struct {
	baseWrapper    *wrappers.BaseWrapper
	defaultTimeout time.Duration
	apiUrl         string
	serviceName    string
}

type IGoTokenomicsWrapper interface {
	GetUsersTokenomicsInfo(userIds []int64, filters []Filter, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersTokenomicsInfoResponseChan
}

func NewGoTokenomicsWrapper(config boilerplate.WrapperConfig) IGoTokenomicsWrapper {
	timeout := 5 * time.Second

	if config.TimeoutSec > 0 {
		timeout = time.Duration(config.TimeoutSec) * time.Second
	}

	if len(config.ApiUrl) == 0 {
		config.ApiUrl = "http://go-tokenomics"

		log.Warn().Msgf("Api Url is missing for GoTokenomics. Setting as default : %v", config.ApiUrl)
	}

	return &Wrapper{
		baseWrapper:    wrappers.GetBaseWrapper(),
		defaultTimeout: timeout,
		apiUrl:         fmt.Sprintf("%v/rpc-service", common.StripSlashFromUrl(config.ApiUrl)),
		serviceName:    "go tokenomics",
	}
}

func (w *Wrapper) GetUsersTokenomicsInfo(userIds []int64, filters []Filter, apmTransaction *apm.Transaction, forceLog bool) chan GetUsersTokenomicsInfoResponseChan {
	respCh := make(chan GetUsersTokenomicsInfoResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "GetUsersTokenomicsInfo", GetUsersTokenomicsInfoRequest{
		UserIds: userIds,
		Filters: filters,
	}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	go func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := GetUsersTokenomicsInfoResponseChan{
			Error: resp.Error,
		}
		respCh <- result
	}()

	return respCh
}
