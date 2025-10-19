package application

import (
	"context"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/pkg/errors"
)

var RecoverFunc = func(ctx context.Context) {
	if er := recover(); er != nil {
		var err error

		switch x := er.(type) {
		case string:
			err = errors.New(x)
		case error:
			err = errors.WithStack(x)
		default:
			// Fallback err (per specs, error strings should be lowercase w/o punctuation
			err = errors.New("unknown panic")
		}

		apm_helper.LogError(err, ctx)
	}
}
