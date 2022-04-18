package apm_helper

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
	"go.elastic.co/apm"
	"go.elastic.co/apm/module/apmhttp"
	"net/http"
	"strconv"
	"time"
)

func init() {
	boilerplate.SetupZeroLog()
}

func StartNewApmTransaction(methodName string, transactionType string, request interface{}, parentTx *apm.Transaction) *apm.Transaction {
	traceContext := apm.TraceContext{}

	if parentTx != nil {
		traceContext = parentTx.TraceContext()
	}

	transaction := apm.DefaultTracer.StartTransactionOptions(methodName, transactionType,
		apm.TransactionOptions{
			TraceContext: traceContext,
			Start:        time.Now(),
		})

	if request != nil {
		AddApmData(transaction, "request", request)
	}

	return transaction
}

func StartNewApmTransactionWithTraceData(methodName string, transactionType string, request interface{}, parentCtx apm.TraceContext) *apm.Transaction {
	transaction := apm.DefaultTracer.StartTransactionOptions(methodName, transactionType,
		apm.TransactionOptions{
			TraceContext: parentCtx,
			Start:        time.Now(),
		})

	if request != nil {
		AddApmData(transaction, "request", request)
	}

	return transaction
}

func AppendRequestBody(request interface{}, transaction *apm.Transaction) {
	if transaction == nil || request == nil {
		return
	}

	AddApmData(transaction, "request", request)
}

func AppendResponseBody(response interface{}, transaction *apm.Transaction) {
	if response == nil || transaction == nil {
		return
	}

	AddApmData(transaction, "response", response)
}

func AppendRequestBodyWithContext(request interface{}, ctx context.Context) {
	AppendRequestBody(request, apm.TransactionFromContext(ctx))
}

func AppendResponseBodyWithContext(response interface{}, ctx context.Context) {
	AppendResponseBody(response, apm.TransactionFromContext(ctx))
}

func LogError(err error, ctx context.Context) {
	if err == nil {
		return
	}

	defer func() {
		_ = recover()
	}()

	log.Ctx(ctx).Err(err).Send()

	apmError := apm.CaptureError(ctx, err)

	if err != nil {
		apmError.Send()
	}
}

func AddApmLabelWithContext(ctx context.Context, key string, value string) {
	AddApmLabel(apm.TransactionFromContext(ctx), key, value)
}

func AddApmLabel(transaction *apm.Transaction, key string, value interface{}) {
	if transaction == nil || len(key) == 0 || value == nil {
		return
	}

	resultStr := stringify(value)

	if len(resultStr) == 0 {
		return
	}

	if len(resultStr) > 1022 { // limit is 1024
		resultStr = resultStr[:1022]
	}

	if transaction.TransactionData != nil {
		transaction.Context.SetLabel(key, resultStr)
	}
}

func AddSpanApmLabelWithContext(ctx context.Context, key string, value string) {
	AddSpanApmLabel(apm.SpanFromContext(ctx), key, value)
}

func AddSpanApmLabel(span *apm.Span, key string, value string) {
	if span == nil || span.Dropped() || span.SpanData == nil {
		return
	}

	resultStr := stringify(value)

	if len(resultStr) == 0 {
		return
	}

	if len(resultStr) > 1022 { // limit is 1024
		resultStr = resultStr[:1022]
	}

	if span.SpanData != nil {
		span.Context.SetLabel(key, resultStr)
	}
}

func AddApmDataWithContext(ctx context.Context, key string, value interface{}) {
	AddApmData(apm.TransactionFromContext(ctx), key, value)
}

func AddApmData(transaction *apm.Transaction, key string, value interface{}) {
	if transaction == nil || len(key) == 0 || value == nil {
		return
	}

	resultStr := stringify(value)

	if len(resultStr) > 0 && transaction.TransactionData != nil {
		transaction.Context.SetCustom(key, resultStr)
	}
}

func stringify(value interface{}) string {
	if value == nil {
		return ""
	}

	switch val := value.(type) {
	case int:
		return strconv.Itoa(val)
	case error:
		return fmt.Sprintf("%+v", val)
	case string:
		return val
	case uint64:
		return strconv.FormatUint(val, 10)
	case bool:
		return strconv.FormatBool(val)
	case []byte:
		return string(val)
	default:
		if data, err := json.Marshal(value); err == nil {
			return string(data)
		}
	}

	return ""
}

func AddDataToSpanTrance(rqSpan *apm.Span, req *fasthttp.Request, ctx context.Context) {
	if rqSpan != nil && !rqSpan.Dropped() {
		r, err := http.NewRequest(
			string(req.Header.Method()),
			string(req.URI().FullURI()), nil)

		if err != nil {
			LogError(err, ctx)
		} else {
			rqSpan.Context.SetHTTPRequest(r)
		}

		req.Header.Set(apmhttp.W3CTraceparentHeader, apmhttp.FormatTraceparentHeader(rqSpan.TraceContext()))

		rqSpan.Context.SetTag("path", string(req.URI().Path()))
		rqSpan.Context.SetTag("full_url", string(req.URI().FullURI()))
	}
}
