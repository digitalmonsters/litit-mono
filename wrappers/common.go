package wrappers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/nodejs"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
	"go.elastic.co/apm"
	"go.elastic.co/apm/module/apmhttp"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type ContentEncodingType string

const (
	ContentEncodingGzip    ContentEncodingType = "gzip"
	ContentEncodingBrotli  ContentEncodingType = "br"
	ContentEncodingDeflate ContentEncodingType = "deflate"
)

type BaseWrapper struct {
	client   *fasthttp.Client
	hostName string
}

var mutex sync.Mutex
var baseWrapper *BaseWrapper

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
		client:   &fasthttp.Client{},
		hostName: hostName,
	}

	return baseWrapper
}

func UnpackFastHttpBody(response *fasthttp.Response) ([]byte, error) {
	encoding := response.Header.Peek("Content-Encoding")

	if len(encoding) == 0 {
		b := make([]byte, len(response.Body()))
		copy(b, response.Body())

		return b, nil
	}

	var err error
	var buf bytes.Buffer
	encodingStr := ContentEncodingType(strings.ToLower(string(encoding)))

	switch encodingStr {
	case ContentEncodingGzip:
		_, err = fasthttp.WriteGunzip(&buf, response.Body())
	case ContentEncodingBrotli:
		_, err = fasthttp.WriteUnbrotli(&buf, response.Body())
	case ContentEncodingDeflate:
		_, err = fasthttp.WriteInflate(&buf, response.Body())
	default:
		err = errors.New(fmt.Sprintf("Cannot decompress response. Unknown Content-Encoding value: %s", encodingStr))
	}

	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
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

func (b *BaseWrapper) SendRequestWithRpcResponseFromAnyService(url string, httpMethod string, contentType string,
	methodName string, request interface{}, timeout time.Duration, apmTransaction *apm.Transaction,
	externalServiceName string, forceLog bool) chan rpc.RpcResponseInternal {

	return b.GetRpcResponseFromAnyService(
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

func (b *BaseWrapper) addDataToSpanTrance(rqSpan *apm.Span, req *fasthttp.Request, apmTransaction *apm.Transaction) {
	if rqSpan != nil && !rqSpan.Dropped() {
		r, err := http.NewRequest(
			string(req.Header.Method()),
			string(req.URI().FullURI()), nil)

		if err != nil {
			apm_helper.CaptureApmError(err, apmTransaction)
		} else {
			rqSpan.Context.SetHTTPRequest(r)
		}

		req.Header.Set(apmhttp.W3CTraceparentHeader, apmhttp.FormatTraceparentHeader(rqSpan.TraceContext()))

		rqSpan.Context.SetTag("path", string(req.URI().Path()))
		rqSpan.Context.SetTag("full_url", string(req.URI().FullURI()))
	}
}

type httpResponseChan struct {
	rawBodyRequest  []byte
	rawBodyResponse []byte
	error           error
	statusCode      int
	span            *apm.Span
	forceLog        bool
}

func (b *BaseWrapper) sendHttpRequestAsync(ctx context.Context, url string, methodName string, request interface{},
	forceLog bool, timeout time.Duration, contentType string, httpMethod string) chan httpResponseChan {
	resultChan := make(chan httpResponseChan, 2)

	result := httpResponseChan{
		forceLog: forceLog,
	}

	go func() {
		defer func() {
			resultChan <- result
			close(resultChan)
		}()

		apmTransaction := apm.TransactionFromContext(ctx)

		if apmTransaction != nil {
			result.span = apmTransaction.StartSpan(fmt.Sprintf("HTTP [%v] [%v]", url, methodName),
				"rpc_internal", nil)

			ctx = apm.ContextWithSpan(ctx, result.span)
		}

		req := fasthttp.AcquireRequest()
		defer fasthttp.ReleaseRequest(req)
		resp := fasthttp.AcquireResponse()
		defer fasthttp.ReleaseResponse(resp)

		defer func() {
			if resp.StatusCode() != 200 {
				result.forceLog = true
			}

			if result.span != nil && !result.span.Dropped() {
				result.span.Context.SetHTTPStatusCode(resp.StatusCode())
			}
		}()

		req.SetRequestURI(url)
		req.Header.SetMethod(httpMethod)
		req.Header.Set("Accept-Encoding", fmt.Sprintf("%s,%s,%s", ContentEncodingBrotli,
			ContentEncodingGzip, ContentEncodingDeflate))
		req.Header.SetContentType(contentType)

		if request != nil {
			if data, err := json.Marshal(request); err != nil {
				result.error = errors.WithStack(err)
				result.forceLog = true

				return
			} else {
				result.rawBodyRequest = data

				req.SetBodyRaw(result.rawBodyRequest)
			}
		}

		b.addDataToSpanTrance(result.span, req, apmTransaction)

		err := b.client.DoTimeout(req, resp, timeout)

		result.statusCode = resp.StatusCode()
		rawBodyResponse, err2 := UnpackFastHttpBody(resp)

		if err2 != nil {
			result.forceLog = true

			if err := apm.CaptureError(ctx, err); err != nil {
				log.Err(err).Send()
			}
		}

		result.rawBodyResponse = rawBodyResponse

		if err != nil {
			result.forceLog = true

			result.error = errors.Wrap(err, fmt.Sprintf("error during sending request to service [%v]. Err: [%v]. StatusCode: [%v]",
				url,
				err.Error(),
				resp.StatusCode()))

			return
		}
	}()

	return resultChan
}

func (b *BaseWrapper) GetRpcResponse(url string, request interface{}, methodName string, timeout time.Duration,
	apmTransaction *apm.Transaction, externalServiceName string, forceLog bool) chan rpc.RpcResponseInternal {
	responseCh := make(chan rpc.RpcResponseInternal, 2)

	ctx := apm.ContextWithTransaction(context.Background(), apmTransaction)

	go func() {
		apiResponse := <-b.sendHttpRequestAsync(ctx, url, methodName, request, forceLog, timeout,
			"application/json", "POST")

		defer func() {
			close(responseCh)

			endRpcSpan(apiResponse.rawBodyRequest, apiResponse.rawBodyResponse, externalServiceName, apiResponse.span,
				apiResponse.forceLog)
		}()

		if apiResponse.error != nil { // its timeout, or some internal error, not logical error
			code := error_codes.GenericServerError

			if errors.Is(apiResponse.error, fasthttp.ErrTimeout) {
				code = error_codes.GenericTimeoutError
			}

			responseCh <- rpc.RpcResponseInternal{
				Error: &rpc.RpcError{
					Code:        code,
					Message:     apiResponse.error.Error(),
					Stack:       fmt.Sprintf("%+v", apiResponse.error),
					Data:        nil,
					Hostname:    b.hostName,
					ServiceName: externalServiceName,
				},
			}

			return
		}

		genericResponse := rpc.RpcResponseInternal{}

		if err := json.Unmarshal(apiResponse.rawBodyResponse, &genericResponse); err != nil {
			wrapped := errors.Wrapf(err, "remote server status code [%v] can not unmarshal to rpc response internal",
				apiResponse.statusCode)

			genericResponse.Error = &rpc.RpcError{
				Code:        error_codes.GenericMappingError,
				Message:     wrapped.Error(),
				Stack:       fmt.Sprintf("%+v", wrapped),
				Data:        nil,
				Hostname:    b.hostName,
				ServiceName: externalServiceName,
			}

			apiResponse.forceLog = true
		}

		if genericResponse.Error != nil {
			apiResponse.forceLog = true

			genericResponse.Result = nil

			genericResponse.Error.Message = fmt.Sprintf("remote server [%v] returned rpc error. [%v]", externalServiceName,
				genericResponse.Error.Message)
		}

		responseCh <- genericResponse
	}()

	return responseCh
}

func (b *BaseWrapper) GetRpcResponseFromNodeJsService(url string, request interface{}, httpMethod string,
	contentType string, methodName string, timeout time.Duration, apmTransaction *apm.Transaction,
	externalServiceName string, forceLog bool) chan rpc.RpcResponseInternal {
	responseCh := make(chan rpc.RpcResponseInternal, 2)

	ctx := apm.ContextWithTransaction(context.Background(), apmTransaction)

	go func() {
		apiResponse := <-b.sendHttpRequestAsync(ctx, url, methodName, request, forceLog, timeout, contentType,
			httpMethod)

		defer func() {
			close(responseCh)

			endRpcSpan(apiResponse.rawBodyRequest, apiResponse.rawBodyResponse, externalServiceName, apiResponse.span,
				apiResponse.forceLog)
		}()

		if apiResponse.error != nil { // its timeout, or some internal error, not logical error
			code := error_codes.GenericServerError

			if errors.Is(apiResponse.error, fasthttp.ErrTimeout) {
				code = error_codes.GenericTimeoutError
			}

			responseCh <- rpc.RpcResponseInternal{
				Error: &rpc.RpcError{
					Code:        code,
					Message:     apiResponse.error.Error(),
					Stack:       fmt.Sprintf("%+v", apiResponse.error),
					Data:        nil,
					Hostname:    b.hostName,
					ServiceName: externalServiceName,
				},
			}

			return
		}

		nodeJsResponse := nodejs.Response{}
		genericResponse := rpc.RpcResponseInternal{}

		if err := json.Unmarshal(apiResponse.rawBodyResponse, &nodeJsResponse); err != nil {
			wrapped := errors.Wrapf(err, "remote server status code [%v] can not unmarshal to nodejs response",
				apiResponse.statusCode)

			genericResponse.Error = &rpc.RpcError{
				Code:        error_codes.GenericMappingError,
				Message:     wrapped.Error(),
				Stack:       fmt.Sprintf("%+v", wrapped),
				Data:        nil,
				Hostname:    b.hostName,
				ServiceName: externalServiceName,
			}

			apiResponse.forceLog = true

			responseCh <- genericResponse

			return
		}

		if !nodeJsResponse.Success {
			apiResponse.forceLog = true

			if nodeJsResponse.Error != nil {
				genericResponse.Error = &rpc.RpcError{
					Code: error_codes.ErrorCode(apiResponse.statusCode),
					Message: errors.New(fmt.Sprintf("remote server [%v] replied with status: [%v] and error: [%v]", externalServiceName,
						nodeJsResponse.Error.Status, nodeJsResponse.Error.Message)).Error(),
					Data:        nil,
					Hostname:    b.hostName,
					ServiceName: externalServiceName,
				}
			} else {
				genericResponse.Error = &rpc.RpcError{
					Code:        error_codes.ErrorCode(apiResponse.statusCode),
					Message:     errors.New("unknown error").Error(),
					Data:        nil,
					Hostname:    b.hostName,
					ServiceName: externalServiceName,
				}
			}

			responseCh <- genericResponse

			return
		}

		genericResponse.Result = nodeJsResponse.Data

		responseCh <- genericResponse
	}()

	return responseCh
}

func (b *BaseWrapper) GetRpcResponseFromAnyService(url string, request interface{}, httpMethod string,
	contentType string, methodName string, timeout time.Duration, apmTransaction *apm.Transaction,
	externalServiceName string, forceLog bool) chan rpc.RpcResponseInternal {
	responseCh := make(chan rpc.RpcResponseInternal, 2)

	ctx := apm.ContextWithTransaction(context.Background(), apmTransaction)

	go func() {
		apiResponse := <-b.sendHttpRequestAsync(ctx, url, methodName, request, forceLog, timeout, contentType,
			httpMethod)

		defer func() {
			close(responseCh)

			endRpcSpan(apiResponse.rawBodyRequest, apiResponse.rawBodyResponse, externalServiceName, apiResponse.span,
				apiResponse.forceLog)
		}()

		if apiResponse.error != nil { // its timeout, or some internal error, not logical error
			code := error_codes.GenericServerError

			if errors.Is(apiResponse.error, fasthttp.ErrTimeout) {
				code = error_codes.GenericTimeoutError
			}

			responseCh <- rpc.RpcResponseInternal{
				Error: &rpc.RpcError{
					Code:        code,
					Message:     apiResponse.error.Error(),
					Stack:       fmt.Sprintf("%+v", apiResponse.error),
					Data:        nil,
					Hostname:    b.hostName,
					ServiceName: externalServiceName,
				},
			}

			return
		}

		unknownResponse := json.RawMessage{}
		genericResponse := rpc.RpcResponseInternal{}

		if err := json.Unmarshal(apiResponse.rawBodyResponse, &unknownResponse); err != nil {
			apiResponse.forceLog = true
			wrapped := errors.Wrapf(err, "remote server status code [%v] can not unmarshal to raw message",
				apiResponse.statusCode)

			genericResponse.Error = &rpc.RpcError{
				Code:        error_codes.GenericMappingError,
				Message:     wrapped.Error(),
				Stack:       fmt.Sprintf("%+v", wrapped),
				Data:        nil,
				Hostname:    b.hostName,
				ServiceName: externalServiceName,
			}

			responseCh <- genericResponse

			return
		}

		genericResponse.Result = unknownResponse

		responseCh <- genericResponse
	}()

	return responseCh
}

func endRpcSpan(rawBodyRequest []byte, rawBodyResponse []byte,
	externalServiceName string, rqSpan *apm.Span, forceLog bool) {
	if rqSpan == nil {
		return
	}

	shouldLog := forceLog

	finalStatement := ""

	if shouldLog && rqSpan != nil {
		if data, err := json.Marshal(map[string]interface{}{
			"request":  rawBodyRequest,
			"response": rawBodyResponse,
		}); err != nil {
			log.Err(err).Send()

			finalStatement = fmt.Sprintf("request [%v] || response [%v]", rawBodyRequest, rawBodyResponse)
		} else {
			finalStatement = string(data)
		}
	}

	rqSpan.Context.SetDatabase(apm.DatabaseSpanContext{
		Instance:  externalServiceName,
		Type:      externalServiceName,
		Statement: finalStatement,
	})

	rqSpan.End()
}
