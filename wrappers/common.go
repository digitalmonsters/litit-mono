package wrappers

import (
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/gammazero/workerpool"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
	"go.elastic.co/apm"
	"strings"
	"sync"
	"time"
)

type BaseWrapper struct {
	workerPool *workerpool.WorkerPool
	client     *fasthttp.Client
}

var mutex sync.Mutex
var baseWrapper *BaseWrapper

func (b *BaseWrapper) GetPool() *workerpool.WorkerPool {
	return b.workerPool
}


func GetBaseWrapper() *BaseWrapper {
	if baseWrapper != nil {
		return baseWrapper
	}

	mutex.Lock()
	defer mutex.Unlock()

	if baseWrapper != nil {
		return baseWrapper
	}

	baseWrapper = &BaseWrapper{workerPool: workerpool.New(1024), client: &fasthttp.Client{}}

	return baseWrapper
}

func (b *BaseWrapper) SendRequestWithRpcResponse(url string, methodName string, request interface{}, timeout time.Duration,
	apmTransaction *apm.Transaction, externalServiceName string, forceLog bool) chan rpc.RpcResponseInternal {

	return b.GetRpcResponse(url, request, methodName, timeout, apmTransaction, externalServiceName, forceLog)
}

func (b *BaseWrapper) SendRpcRequest(url string, methodName string, request interface{}, timeout time.Duration,
	apmTransaction *apm.Transaction, externalServiceName string, forceLog bool) chan rpc.RpcResponseInternal {
	name := strings.ToLower(methodName)
	return b.GetRpcResponse(url, rpc.RpcRequestInternal{
		Method:  name,
		Params:  request,
		Id:      "1",
		JsonRpc: "2.0",
	}, name, timeout, apmTransaction, externalServiceName, forceLog)
}

func (b *BaseWrapper) GetRpcResponse(url string, request interface{}, methodName string, timeout time.Duration,
	apmTransaction *apm.Transaction, externalServiceName string, forceLog bool) chan rpc.RpcResponseInternal {
	responseCh := make(chan rpc.RpcResponseInternal, 2)

	b.workerPool.Submit(func() {
		defer func() {
			close(responseCh)
		}()

		var rqTransaction *apm.Transaction
		var rawBodyRequest []byte
		var rawBodyResponse []byte
		var genericResponse *rpc.RpcResponseInternal

		if apmTransaction != nil {
			rqTransaction = apm_helper.StartNewApmTransaction(fmt.Sprintf("HTTP [%v] [%v]", url, methodName),
				"internal_rpc", nil, apmTransaction)
		}

		defer func() {
			if rqTransaction == nil {
				return
			}

			shouldLog := forceLog

			if genericResponse != nil && genericResponse.Error != nil {
				shouldLog = true // we have an error

				apm_helper.CaptureApmError(errors.New(fmt.Sprintf("external service [%v] replay with error [%v] and msg [%v]",
					externalServiceName, genericResponse.Error.Code, genericResponse.Error.Message)), rqTransaction)
			}

			if shouldLog {
				apm_helper.AddApmData(rqTransaction, "raw_request", rawBodyRequest)
				apm_helper.AddApmData(rqTransaction, "raw_response", rawBodyResponse)
				apm_helper.AddApmData(rqTransaction, "parsed_response", genericResponse)
			}

			rqTransaction.End()
		}()

		req := fasthttp.AcquireRequest()
		defer fasthttp.ReleaseRequest(req)
		resp := fasthttp.AcquireResponse()
		defer fasthttp.ReleaseResponse(resp)

		req.SetRequestURI(url)
		req.Header.SetMethod("POST")

		if request != nil {
			if b, err := json.Marshal(request); err != nil {
				genericResponse = &rpc.RpcResponseInternal{
					Error: &rpc.RpcError{
						Code:    error_codes.GenericMappingError,
						Message: err.Error(),
						Data:    nil,
					},
				}

				responseCh <- *genericResponse
				return
			} else {
				rawBodyRequest = b

				req.SetBodyRaw(rawBodyRequest)
			}
		}

		if err := b.client.DoTimeout(req, resp, timeout); err != nil {
			code := error_codes.GenericServerError

			if errors.Is(err, fasthttp.ErrTimeout) {
				code = error_codes.GenericTimeoutError
			}

			genericResponse = &rpc.RpcResponseInternal{
				Error: &rpc.RpcError{
					Code:    code,
					Message: "error during sending request",
				},
			}

			responseCh <- *genericResponse

			return
		}

		rawBodyResponse = resp.Body()

		genericResponse = &rpc.RpcResponseInternal{}

		if err := json.Unmarshal(rawBodyResponse, genericResponse); err != nil {
			genericResponse.Error = &rpc.RpcError{
				Code:    error_codes.GenericMappingError,
				Message: err.Error(),
				Data:    nil,
			}

			responseCh <- *genericResponse

			return
		}

		responseCh <- *genericResponse
	})

	return responseCh
}
