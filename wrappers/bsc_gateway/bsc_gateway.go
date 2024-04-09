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
	"go.elastic.co/apm"
)

type IBscGatewayWrapper interface {
	CreateSignature(amount string, withdrawalTransactionId, userId, adminId int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[SignatureResponseData]
	GetSignatures(withdrawalIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]SignatureResponseData]
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

func (w BscGatewayWrapper) CreateSignature(amount string,
	withdrawalTransactionId, userId, adminId int64,
	ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[SignatureResponseData] {

	return wrappers.ExecuteRpcRequestAsync[SignatureResponseData](w.baseWrapper, w.apiUrl, "CreateSignature", SignatureRequest{
		Amount:                  amount,
		WithdrawalTransactionId: withdrawalTransactionId,
		UserId:                  userId,
		AdminId:                 adminId,
	}, map[string]string{}, w.defaultTimeout, apm.TransactionFromContext(ctx), w.serviceName, forceLog)
}

func (w BscGatewayWrapper) GetSignatures(withdrawalIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]SignatureResponseData] {
	return wrappers.ExecuteRpcRequestAsync[map[int64]SignatureResponseData](w.baseWrapper, w.apiUrl, "GetSignatures",
		GetSignaturesRequest{WithdrawalIds: withdrawalIds}, map[string]string{}, w.defaultTimeout, apm.TransactionFromContext(ctx), w.serviceName, forceLog)
}
