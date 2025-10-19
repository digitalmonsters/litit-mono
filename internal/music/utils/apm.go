package utils

import (
	"context"
	"github.com/digitalmonsters/go-common/apm_helper"
	"go.elastic.co/apm"
)

func CaptureApmErrorFromTransaction(err error, ctx context.Context) {
	if err == nil {
		return
	}

	transaction := apm.TransactionFromContext(ctx)

	if transaction == nil || transaction.TransactionData == nil {
		return
	}

	exx := apm.DefaultTracer.NewError(err)

	apm_helper.AddApmData(transaction, "exception", err)

	exx.SetTransaction(transaction)
	exx.Context = transaction.Context

	exx.Send()
}
