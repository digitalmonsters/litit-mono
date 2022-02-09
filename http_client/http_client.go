package http_client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/common"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
	"go.elastic.co/apm"
	urlPackage "net/url"
	"time"
)

var defaultClient = fasthttp.Client{}

type HttpResult struct {
	rawRequest  []byte
	rawResponse []byte
	error       error
	forceLog    bool
	statusCode  int
}

func (h HttpResult) GetRawRequest() []byte {
	return h.rawRequest
}

func (h HttpResult) GetRawResponse() []byte {
	return h.rawResponse
}

func (h HttpResult) GetError() error {
	return h.error
}

func (h HttpResult) GetStatusCode() int {
	return h.statusCode
}

func endRpcSpan(rawBodyRequest []byte, rawBodyResponse []byte,
	rqSpan *apm.Span, forceLog bool, host string) {
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
		Instance:  host,
		Type:      host,
		Statement: finalStatement,
	})

	rqSpan.End()
}

func SendHttpRequestAsync(ctx context.Context, url string, methodName string, contentType string, httpMethod string,
	request interface{}, forceLog bool, timeout time.Duration) chan HttpResult {
	return sendHttpRequestAsync(ctx, url, methodName, contentType, httpMethod, request, forceLog, timeout, false, 0)
}

func SendHttpRequestAsyncWithRedirects(ctx context.Context, url string, methodName string, contentType string, httpMethod string,
	request interface{}, forceLog bool, timeout time.Duration, maxRedirectCount int) chan HttpResult {
	return sendHttpRequestAsync(ctx, url, methodName, contentType, httpMethod, request, forceLog, timeout, true, maxRedirectCount)
}

func sendHttpRequestAsync(ctx context.Context, url string, methodName string, contentType string, httpMethod string,
	request interface{}, forceLog bool, timeout time.Duration, allowedRedirects bool, maxRedirectCount int) chan HttpResult {
	ch := make(chan HttpResult, 2)

	go func() {
		result := HttpResult{
			forceLog: forceLog,
		}

		pasedUrl, err := urlPackage.Parse(url)

		if err != nil {
			result.error = errors.WithStack(err)
			return
		}

		ctx, cancel := context.WithTimeout(ctx, timeout)

		apmTransaction := apm.TransactionFromContext(ctx)

		var span *apm.Span

		if apmTransaction != nil {
			span = apmTransaction.StartSpan(fmt.Sprintf("HTTP [%v] [%v]", url, methodName),
				"rpc_internal", nil)

			ctx = apm.ContextWithSpan(ctx, span)
		}

		defer func() {
			cancel()
			ch <- result

			if span != nil {
				span.Context.SetHTTPStatusCode(result.statusCode)
				endRpcSpan(result.rawRequest, result.rawResponse, span,
					result.forceLog, pasedUrl.Hostname())
			}

			close(ch)
		}()

		req := fasthttp.AcquireRequest()
		defer fasthttp.ReleaseRequest(req)
		resp := fasthttp.AcquireResponse()
		defer fasthttp.ReleaseResponse(resp)

		req.SetRequestURI(url)
		req.Header.SetMethod(httpMethod)
		req.Header.Set("Accept-Encoding", fmt.Sprintf("%s,%s,%s", common.ContentEncodingBrotli,
			common.ContentEncodingGzip, common.ContentEncodingDeflate))
		req.Header.SetContentType(contentType)

		if request != nil {
			if data, err := json.Marshal(request); err != nil {
				result.error = errors.WithStack(err)
				result.forceLog = true

				return
			} else {
				result.rawRequest = data

				req.SetBodyRaw(result.rawRequest)
			}
		}

		apm_helper.AddDataToSpanTrance(span, req, apmTransaction)

		if allowedRedirects{
			err = defaultClient.DoRedirects(req, resp, maxRedirectCount)
		}else{
			err = defaultClient.DoTimeout(req, resp, timeout)
		}
		result.statusCode = resp.StatusCode()

		rawBodyResponse, err2 := common.UnpackFastHttpBody(resp)

		if err2 != nil {
			result.forceLog = true

			if err := apm.CaptureError(ctx, err); err != nil {
				log.Err(err).Send()
			}
		}

		result.rawResponse = rawBodyResponse

		if err != nil {
			result.forceLog = true

			result.error = errors.Wrap(err, fmt.Sprintf("error during sending request to [%v]. Err: [%v]. StatusCode: [%v]",
				url,
				err.Error(),
				resp.StatusCode()))

			return
		}
	}()

	return ch
}
