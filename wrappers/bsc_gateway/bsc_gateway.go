package bsc_gateway

import (
	"context"
	"fmt"
	"time"

	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/wrappers"
	"github.com/patrickmn/go-cache"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
	"go.elastic.co/apm"
)

type IBscGatewayWrapper interface {
	CreateSignature(from string, amount decimal.Decimal, withdrawalTransactionId, userId, adminId int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[SignatureResponseData]
	GetSignatureStatus(withdrawalId int64, signature string, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[SignatureStatusResponseData]
}

type BscGatewayWrapper struct {
	baseWrapper    *wrappers.BaseWrapper
	defaultTimeout time.Duration
	apiUrl         string
	serviceName    string
	cache          *cache.Cache
}

func NewBscGatewayWrapper(config boilerplate.WrapperConfig) IBscGatewayWrapper {
	timeout := 25 * time.Second

	if config.TimeoutSec > 0 {
		timeout = time.Duration(config.TimeoutSec) * time.Second
	}

	if len(config.ApiUrl) == 0 {
		config.ApiUrl = "http://bsc-gateway"

		log.Warn().Msgf("Api Url is missing for BscGateway. Setting as default : %v", config.ApiUrl)
	}

	return &BscGatewayWrapper{
		baseWrapper:    wrappers.GetBaseWrapper(),
		defaultTimeout: timeout,
		apiUrl:         fmt.Sprintf("%v/rpc-service", common.StripSlashFromUrl(config.ApiUrl)),
		serviceName:    "bsc-gateway",
		cache:          cache.New(4*time.Minute, 5*time.Minute),
	}
}

func (w BscGatewayWrapper) CreateSignature(
	from string, amount decimal.Decimal,
	withdrawalTransactionId, userId, adminId int64,
	ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[SignatureResponseData] {

	return wrappers.ExecuteRpcRequestAsync[SignatureResponseData](w.baseWrapper, w.apiUrl, "CreateSignature", CreateSignatureRequest{
		From:                    from,
		Amount:                  amount,
		WithdrawalTransactionId: withdrawalTransactionId,
		UserId:                  userId,
		AdminId:                 adminId,
	}, map[string]string{}, w.defaultTimeout, apm.TransactionFromContext(ctx), w.serviceName, forceLog)
}

func (w BscGatewayWrapper) GetSignatureStatus(withdrawalId int64, signature string, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[SignatureStatusResponseData] {
	return wrappers.ExecuteRpcRequestAsync[SignatureStatusResponseData](w.baseWrapper, w.apiUrl, "GetSignatureStatus",
		GetSignatureStatusRequest{withdrawalId, signature}, map[string]string{}, w.defaultTimeout, apm.TransactionFromContext(ctx), w.serviceName, forceLog)
}
