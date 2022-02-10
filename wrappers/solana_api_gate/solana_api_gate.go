package solana_api_gate

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

type ISolanaApiGateWrapper interface {
	TransferToken(from string, amount string, account string, recipientType string, apmTransaction *apm.Transaction, forceLog bool) chan TransferTokenResponseChan
	CreateVesting(from string, to string, amounts string, timestamps string, apmTransaction *apm.Transaction, forceLog bool) chan CreateVestingResponseChan
}

//goland:noinspection GoNameStartsWithPackageName
type SolanaApiGateWrapper struct {
	baseWrapper    *wrappers.BaseWrapper
	defaultTimeout time.Duration
	apiUrl         string
	serviceName    string
	cache          *cache.Cache
}

func NewSolanaApiGateWrapper(config boilerplate.WrapperConfig) ISolanaApiGateWrapper {
	timeout := 5 * time.Second

	if config.TimeoutSec > 0 {
		timeout = time.Duration(config.TimeoutSec) * time.Second
	}

	if len(config.ApiUrl) == 0 {
		config.ApiUrl = "http://solana-api-gate"

		log.Warn().Msgf("Api Url is missing for SolanaApiGate. Setting as default : %v", config.ApiUrl)
	}

	return &SolanaApiGateWrapper{
		baseWrapper:    wrappers.GetBaseWrapper(),
		defaultTimeout: timeout,
		apiUrl:         fmt.Sprintf("%v/rpc-service", common.StripSlashFromUrl(config.ApiUrl)),
		serviceName:    "user-go",
		cache:          cache.New(4*time.Minute, 5*time.Minute),
	}
}

func (w SolanaApiGateWrapper) TransferToken(from string, amount string, account string, recipientType string, apmTransaction *apm.Transaction, forceLog bool) chan TransferTokenResponseChan {
	respCh := make(chan TransferTokenResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "TransferToken", TransferRequest{
		From:   from,
		Amount: amount,
		To: &Recipient{
			Account: account,
			Type:    recipientType,
		},
	},map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	go func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := TransferTokenResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			var transferTokenResponse TransferTokenResponseData

			if err := json.Unmarshal(resp.Result, &transferTokenResponse); err != nil {
				result.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    w.baseWrapper.GetHostName(),
					ServiceName: w.serviceName,
				}
			} else {
				result.Data = &transferTokenResponse
			}
		}

		respCh <- result
	}()

	return respCh
}


func (w SolanaApiGateWrapper) CreateVesting(from string, to string, amounts string, timestamps string, apmTransaction *apm.Transaction, forceLog bool) chan CreateVestingResponseChan {
	respCh := make(chan CreateVestingResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "CreateVesting", CreateVestingRequest{
		From:       from,
		To:         to,
		Amounts:    amounts,
		Timestamps: timestamps,
	}, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	go func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := CreateVestingResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			var createVestingResponse CreateVestingResponseData

			if err := json.Unmarshal(resp.Result, &createVestingResponse); err != nil {
				result.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    w.baseWrapper.GetHostName(),
					ServiceName: w.serviceName,
				}
			} else {
				result.Data = &createVestingResponse
			}
		}

		respCh <- result
	}()

	return respCh
}