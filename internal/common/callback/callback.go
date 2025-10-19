package callback

import (
	"context"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/application"
)

type Callback func(ctx context.Context) error

func ExecuteCallbacksAsync(ctx context.Context, callback ...Callback) {
	go func() {
		ExecuteCallbacks(ctx, callback...)
	}()
}

func ExecuteCallbacks(ctx context.Context, callback ...Callback) {
	if len(callback) == 0 {
		return
	}

	for _, c := range callback {
		func() {
			defer application.RecoverFunc(ctx)

			if err := c(ctx); err != nil {
				apm_helper.LogError(err, ctx)
			}
		}()
	}
}
