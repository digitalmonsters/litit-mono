package content_uploader

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/http_client"
	"github.com/digitalmonsters/go-common/wrappers"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"go.elastic.co/apm"
	"time"
)

type IContentUploaderWrapper interface {
	UploadContentInternal(url string, contentType string, data []byte, apmTransaction *apm.Transaction, forceLog bool) chan error
}

type ContentUploaderWrapper struct {
	defaultTimeout time.Duration
	apiUrl         string
	serviceName    string
	baseWrapper    *wrappers.BaseWrapper
}

func NewContentUploaderWrapper(config boilerplate.WrapperConfig) IContentUploaderWrapper {
	timeout := 15 * time.Second

	if config.TimeoutSec > 0 {
		timeout = time.Duration(config.TimeoutSec) * time.Second
	}

	if len(config.ApiUrl) == 0 {
		config.ApiUrl = "http://content-uploader"

		log.Warn().Msgf("Api Url is missing for ContentUploader. Setting as default : %v", config.ApiUrl)
	}

	return &ContentUploaderWrapper{
		defaultTimeout: timeout,
		apiUrl:         common.StripSlashFromUrl(config.ApiUrl),
		serviceName:    "content_uploader",
		baseWrapper:    wrappers.GetBaseWrapper(),
	}
}

func (w ContentUploaderWrapper) UploadContentInternal(path string, contentType string, data []byte, apmTransaction *apm.Transaction, forceLog bool) chan error {
	resChan := make(chan error, 2)

	go func() {
		defer close(resChan)
		ctx := apm.ContextWithTransaction(context.Background(), apmTransaction)
		resp, err := http_client.DefaultHttpClient.
			NewRequest(ctx).
			SetContentType(contentType).
			SetBodyBytes(data).
			Put(fmt.Sprintf("%v/%v", w.apiUrl, path))
		if err != nil {
			resChan <- err
		}
		if !resp.IsSuccess() {
			resChan <- errors.New(fmt.Sprint(resp.Error()))
		}
		resChan <- nil
	}()

	return resChan
}
