package go_tokenomics

import (
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/digitalmonsters/go-common/wrappers"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
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
	GetWithdrawalsAmountsByAdminIds(adminIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetWithdrawalsAmountsByAdminIdsResponseChan
	GetContentEarningsTotalByContentIds(contentIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetContentEarningsTotalByContentIdsResponseChan
	GetTokenomicsStatsByUserId(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetTokenomicsStatsByUserIdResponseChan
	GetConfigProperties(properties []string, apmTransaction *apm.Transaction, forceLog bool) chan GetConfigPropertiesResponseChan
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
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

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

func (w *Wrapper) GetWithdrawalsAmountsByAdminIds(adminIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetWithdrawalsAmountsByAdminIdsResponseChan {
	respCh := make(chan GetWithdrawalsAmountsByAdminIdsResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "GetWithdrawalsAmountsByAdminIds", GetWithdrawalsAmountsByAdminIdsRequest{
		AdminIds: adminIds,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	go func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := GetWithdrawalsAmountsByAdminIdsResponseChan{
			Error: resp.Error,
		}

		respCh <- result
	}()

	return respCh
}

func (w *Wrapper) GetContentEarningsTotalByContentIds(contentIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetContentEarningsTotalByContentIdsResponseChan {
	respCh := make(chan GetContentEarningsTotalByContentIdsResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "GetContentEarningsTotal", GetContentEarningsTotalByContentIdsRequest{
		ContentIds: contentIds,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	go func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := GetContentEarningsTotalByContentIdsResponseChan{
			Error: resp.Error,
		}
		if len(resp.Result) > 0 {
			data := map[int64]decimal.Decimal{}

			if err := json.Unmarshal(resp.Result, &data); err != nil {
				result.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    w.baseWrapper.GetHostName(),
					ServiceName: w.serviceName,
				}
			} else {
				result.Items = data
			}
		}

		respCh <- result
	}()

	return respCh
}


func (w *Wrapper) GetTokenomicsStatsByUserId(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan GetTokenomicsStatsByUserIdResponseChan {
	respCh := make(chan GetTokenomicsStatsByUserIdResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "GetTokenomicsStatsByUserId", GetTokenomicsStatsByUserIdRequest{
		UserIds: userIds,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	go func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := GetTokenomicsStatsByUserIdResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			var data = make(map[int64]*UserTokenomicsStats)

			if err := json.Unmarshal(resp.Result, &data); err != nil {
				result.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    w.baseWrapper.GetHostName(),
					ServiceName: w.serviceName,
				}
			} else {
				result.Items = data
			}
		}

		respCh <- result
	}()

	return respCh
}

func (w *Wrapper) GetConfigProperties(properties []string, apmTransaction *apm.Transaction, forceLog bool) chan GetConfigPropertiesResponseChan{
	respCh := make(chan GetConfigPropertiesResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "GetConfigProperties", GetConfigPropertiesRequest{Properties: properties},
	map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	go func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := GetConfigPropertiesResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			var data = make(map[string]string)

			if err := json.Unmarshal(resp.Result, &data); err != nil {
				result.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    w.baseWrapper.GetHostName(),
					ServiceName: w.serviceName,
				}
			} else {
				result.Items = data
			}
		}
		respCh <- result
	}()

	return respCh
}