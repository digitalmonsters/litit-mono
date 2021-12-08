package wrappers

import (
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/nodejs"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/gammazero/workerpool"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
	"go.elastic.co/apm"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type BaseWrapper struct {
	workerPool *workerpool.WorkerPool
	client     *fasthttp.Client
	hostName   string
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

	hostName, _ := os.Hostname()

	baseWrapper = &BaseWrapper{
		workerPool: workerpool.New(1024),
		client:     &fasthttp.Client{},
		hostName:   hostName,
	}

	return baseWrapper
}

func (b *BaseWrapper) GetHostName() string {
	return b.hostName
}

func (b *BaseWrapper) SendRequestWithRpcResponse(url string, methodName string, request interface{}, timeout time.Duration,
	apmTransaction *apm.Transaction, externalServiceName string, forceLog bool) chan rpc.RpcResponseInternal {

	return b.GetRpcResponse(url, request, methodName, timeout, apmTransaction, externalServiceName, forceLog)
}

func (b *BaseWrapper) SendRequestWithRpcResponseFromNodeJsService(url string, httpMethod string, contentType string,
	methodName string, request interface{}, timeout time.Duration, apmTransaction *apm.Transaction,
	externalServiceName string, forceLog bool) chan rpc.RpcResponseInternal {

	return b.GetRpcResponseFromNodeJsService(
		url, request, httpMethod, contentType, methodName, timeout, apmTransaction, externalServiceName, forceLog,
	)
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

		//var rqTransaction *apm.Transaction
		var rawBodyRequest []byte
		var rawBodyResponse []byte
		var genericResponse *rpc.RpcResponseInternal
		var rqSpan *apm.Span

		if apmTransaction != nil {
			rqSpan = apmTransaction.StartSpan(fmt.Sprintf("HTTP [%v] [%v]", url, methodName),
				"rpc_internal", nil)
		}

		req := fasthttp.AcquireRequest()
		defer fasthttp.ReleaseRequest(req)
		resp := fasthttp.AcquireResponse()
		defer fasthttp.ReleaseResponse(resp)

		defer func() {
			if rqSpan != nil && !rqSpan.Dropped() {
				rqSpan.Context.SetHTTPStatusCode(resp.StatusCode())
			}

			endRpcTransaction(genericResponse, rawBodyRequest, rawBodyResponse, externalServiceName, rqSpan, forceLog)
		}()

		req.SetRequestURI(url)
		req.Header.SetMethod("POST")

		if request != nil {
			if data, err := json.Marshal(request); err != nil {
				genericResponse = &rpc.RpcResponseInternal{
					Error: &rpc.RpcError{
						Code:     error_codes.GenericMappingError,
						Message:  err.Error(),
						Data:     nil,
						Hostname: b.hostName,
					},
				}

				responseCh <- *genericResponse
				return
			} else {
				rawBodyRequest = data

				req.SetBodyRaw(rawBodyRequest)
			}
		}

		if rqSpan != nil && !rqSpan.Dropped() {
			r, err := http.NewRequest(
				string(req.Header.Method()),
				string(req.URI().FullURI()), nil)

			if err != nil {
				apm_helper.CaptureApmError(err, apmTransaction)
			} else {
				rqSpan.Context.SetHTTPRequest(r)
			}

			rqSpan.Context.SetDestinationService(apm.DestinationServiceSpanContext{
				Name:     externalServiceName,
				Resource: externalServiceName,
			})
		}

		if err := b.client.DoTimeout(req, resp, timeout); err != nil {
			code := error_codes.GenericServerError

			if errors.Is(err, fasthttp.ErrTimeout) {
				code = error_codes.GenericTimeoutError
			}

			rawBodyResponse = resp.Body()

			genericResponse = &rpc.RpcResponseInternal{
				Error: &rpc.RpcError{
					Code:     code,
					Message:  fmt.Sprintf("error during sending request. Remote server status code [%v]", resp.StatusCode()),
					Hostname: b.hostName,
					Data: map[string]interface{}{
						"raw_response": string(rawBodyResponse),
					},
				},
			}

			responseCh <- *genericResponse

			return
		}

		rawBodyResponse = resp.Body()

		genericResponse = &rpc.RpcResponseInternal{}

		if err := json.Unmarshal(rawBodyResponse, genericResponse); err != nil {
			genericResponse.Error = &rpc.RpcError{
				Code:     error_codes.GenericMappingError,
				Message:  err.Error(),
				Data:     nil,
				Hostname: b.hostName,
			}

			responseCh <- *genericResponse

			return
		}

		responseCh <- *genericResponse
	})

	return responseCh
}

func (b *BaseWrapper) GetRpcResponseFromNodeJsService(url string, request interface{}, httpMethod string,
	contentType string, methodName string, timeout time.Duration, apmTransaction *apm.Transaction,
	externalServiceName string, forceLog bool) chan rpc.RpcResponseInternal {
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
			rqTransaction = apm_helper.StartNewApmTransaction(fmt.Sprintf("HTTP [%s] [%v] [%v]", httpMethod, url, methodName),
				"internal_rpc", nil, apmTransaction)
		}

		req := fasthttp.AcquireRequest()
		defer fasthttp.ReleaseRequest(req)
		resp := fasthttp.AcquireResponse()
		defer fasthttp.ReleaseResponse(resp)

		defer func() {
			apm_helper.AddApmLabel(rqTransaction, "remote_status_code", resp.StatusCode())
			endRpcTransaction(genericResponse, rawBodyRequest, rawBodyResponse, externalServiceName, nil, forceLog) // todo
		}()

		apm_helper.AddApmLabel(rqTransaction, "remote_path", url)
		req.SetRequestURI(url)
		req.Header.SetMethod(httpMethod)
		apm_helper.AddApmLabel(rqTransaction, "remote_url", string(req.URI().FullURI()))

		if request != nil {
			if data, err := json.Marshal(request); err != nil {
				genericResponse = &rpc.RpcResponseInternal{
					Error: &rpc.RpcError{
						Code:     error_codes.GenericMappingError,
						Message:  err.Error(),
						Data:     nil,
						Hostname: b.hostName,
					},
				}

				responseCh <- *genericResponse
				return
			} else {
				rawBodyRequest = data

				req.Header.SetContentType(contentType)
				req.SetBodyRaw(rawBodyRequest)
			}
		}

		if err := b.client.DoTimeout(req, resp, timeout); err != nil {
			code := error_codes.GenericServerError

			if errors.Is(err, fasthttp.ErrTimeout) {
				code = error_codes.GenericTimeoutError
			}

			rawBodyResponse = resp.Body()

			genericResponse = &rpc.RpcResponseInternal{
				Error: &rpc.RpcError{
					Code:     code,
					Message:  fmt.Sprintf("error during sending request. Remote server status code [%v]", resp.StatusCode()),
					Hostname: b.hostName,
					Data: map[string]interface{}{
						"raw_response": string(rawBodyResponse),
					},
				},
			}

			responseCh <- *genericResponse

			return
		}

		rawBodyResponse = resp.Body()

		nodeJsResponse := &nodejs.Response{}
		genericResponse = &rpc.RpcResponseInternal{}

		if err := json.Unmarshal(rawBodyResponse, nodeJsResponse); err != nil {
			genericResponse.Error = &rpc.RpcError{
				Code:     error_codes.GenericMappingError,
				Message:  err.Error(),
				Data:     nil,
				Hostname: b.hostName,
			}

			responseCh <- *genericResponse

			return
		}

		if !nodeJsResponse.Success {
			if nodeJsResponse.Error != nil {
				genericResponse.Error = &rpc.RpcError{
					Code:     error_codes.GenericServerError,
					Message:  errors.New(fmt.Sprintf("status: %v, error: %s", nodeJsResponse.Error.Status, nodeJsResponse.Error.Message)).Error(),
					Data:     nil,
					Hostname: b.hostName,
				}
			} else {
				genericResponse.Error = &rpc.RpcError{
					Code:     error_codes.GenericServerError,
					Message:  errors.New("unknown error").Error(),
					Data:     nil,
					Hostname: b.hostName,
				}
			}

			responseCh <- *genericResponse

			return
		}

		genericResponse.Result = nodeJsResponse.Data

		responseCh <- *genericResponse
	})

	return responseCh
}

func endRpcTransaction(genericResponse *rpc.RpcResponseInternal, rawBodyRequest []byte, rawBodyResponse []byte,
	externalServiceName string, rqSpan *apm.Span, forceLog bool) {
	if rqSpan == nil {
		return
	}

	shouldLog := forceLog

	if genericResponse != nil && genericResponse.Error != nil {
		shouldLog = true // we have an error
	}

	if shouldLog && rqSpan != nil {
		if data, err := json.Marshal(map[string]interface{}{
			"raw_response":    rawBodyResponse,
			"raw_request":     rawBodyRequest,
			"parsed_response": genericResponse,
		}); err != nil {
			rqSpan.Context.SetDatabase(apm.DatabaseSpanContext{
				Instance:  externalServiceName,
				Type:      externalServiceName,
				Statement: fmt.Sprintf("can not log due to error in serialization %+v", errors.WithStack(err)),
			})

		} else {
			rqSpan.Context.SetDatabase(apm.DatabaseSpanContext{
				Instance:  externalServiceName,
				Type:      externalServiceName,
				Statement: string(data),
			})
		}
	}

	rqSpan.End()
}
