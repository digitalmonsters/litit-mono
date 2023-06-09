package sqs_listener

import (
	"context"

	"go.elastic.co/apm"
)

type ExecutionData struct {
	ApmTransaction *apm.Transaction
	Context        context.Context
}
