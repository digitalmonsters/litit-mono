package solana_api_gate

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/wrappers"
	"github.com/patrickmn/go-cache"
	"github.com/rs/zerolog/log"
	"go.elastic.co/apm"
	"time"
)

type ISolanaApiGateWrapper interface {
	TransferToken(from string, amount string, account string, recipientType string, withdrawalTransactionId int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[TransferTokenResponseData]
	CreateVesting(from string, to string, amounts string, timestamps string, withdrawalTransactionId int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[CreateVestingResponseData]
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
	timeout := 25 * time.Second

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
		serviceName:    "solana-api-gate",
		cache:          cache.New(4*time.Minute, 5*time.Minute),
	}
}

func (w SolanaApiGateWrapper) TransferToken(from string, amount string, account string, recipientType string, withdrawalTransactionId int64,
	ctx context.Context,
	forceLog bool) chan wrappers.GenericResponseChan[TransferTokenResponseData] {

	return wrappers.ExecuteRpcRequestAsync[TransferTokenResponseData](w.baseWrapper, w.apiUrl, "TransferToken", TransferRequest{
		From:                    from,
		Amount:                  amount,
		WithdrawalTransactionId: withdrawalTransactionId,
		To: &Recipient{
			Account: account,
			Type:    recipientType,
		},
	}, map[string]string{}, w.defaultTimeout, apm.TransactionFromContext(ctx), w.serviceName, forceLog)
}

func (w SolanaApiGateWrapper) CreateVesting(from string, to string, amounts string, timestamps string, withdrawalTransactionId int64,
	ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[CreateVestingResponseData] {

	return wrappers.ExecuteRpcRequestAsync[CreateVestingResponseData](w.baseWrapper, w.apiUrl, "CreateVesting", CreateVestingRequest{
		From:                    from,
		To:                      to,
		Amounts:                 amounts,
		Timestamps:              timestamps,
		WithdrawalTransactionId: withdrawalTransactionId,
	}, map[string]string{}, w.defaultTimeout, apm.TransactionFromContext(ctx), w.serviceName, forceLog)
}
