package apm_helper

import (
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/common"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
	"go.elastic.co/apm"
	"strconv"
	"time"
)

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

func CaptureApmError(err error, transaction *apm.Transaction) {
	if err == nil {
		return
	}

	if transaction == nil || transaction.TransactionData == nil {
		return
	}

	log.Err(err).Send() // temp

	exx := apm.DefaultTracer.NewError(err)

	AddApmData(transaction, "exception", err)

	exx.SetTransaction(transaction)
	exx.Context = transaction.Context

	exx.Send()
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

func SendHttpRequest(client *fasthttp.Client, request *fasthttp.Request, response *fasthttp.Response, parentTx *apm.Transaction,
	timeout time.Duration, logResponse bool) error {
	if request == nil {
		return errors.New("request should not be nil")
	}
	if response == nil {
		return errors.New("response should not be nil")
	}

	tx := StartNewApmTransaction(
		fmt.Sprintf("[%v] %v", string(request.Header.Method()), request.URI().String()),
		"http_external",
		nil,
		parentTx,
	)

	defer tx.End()

	err := client.DoTimeout(request, response, timeout)

	if err != nil {
		return err
	}

	data, _ := common.UnpackFastHttpBody(response)

	tx.Context.SetHTTPStatusCode(response.StatusCode())

	if logResponse || (response.StatusCode() != 200 && response.StatusCode() != 201) {
		headers := map[string]string{}

		response.Header.VisitAll(func(key, value []byte) {
			headers[string(key)] = string(value)
		})

		b, _ := json.Marshal(headers)

		AddApmData(tx, "headers", string(b))
		tx.Context.SetCustom("response_body", string(data))
	}

	return nil
}