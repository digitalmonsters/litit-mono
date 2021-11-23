package apm_helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"go.elastic.co/apm"
	"io/ioutil"
	"net/http"
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

func SendRequest(client *http.Client, req *http.Request, parentTx *apm.Transaction, logResponse bool) (*http.Response, error) {
	tx := StartNewApmTransaction(
		fmt.Sprintf("[%v] %v", req.Method, req.URL.String()),
		"http_external",
		nil,
		parentTx,
	)

	defer tx.End()

	if tx.TransactionData != nil {
		tx.Context.SetHTTPRequest(req)
	}

	if req.Body != nil {
		requestBody, err := req.GetBody()

		if err != nil {
			return nil, errors.WithStack(err)
		}

		if requestBody != nil {
			requestContent, _ := ioutil.ReadAll(requestBody)
			tx.Context.SetCustom("request_body", string(requestContent))
		}
	}

	resp, err := client.Do(req)

	if err != nil {
		return resp, errors.WithStack(err)
	}

	responseBody, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	resp.Body = ioutil.NopCloser(bytes.NewBuffer(responseBody))

	tx.Context.SetHTTPStatusCode(resp.StatusCode)

	if logResponse || resp.StatusCode != 200 {
		tx.Context.SetCustom("response_body", string(responseBody))
	}

	return resp, nil
}
