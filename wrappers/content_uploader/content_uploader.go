package content_uploader

import (
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/digitalmonsters/go-common/wrappers"
	"github.com/rs/zerolog/log"
	"go.elastic.co/apm"
	"time"
)

type IContentUploaderWrapper interface {
	UploadContentInternal(url string, contentType string, data string, apmTransaction *apm.Transaction, forceLog bool) chan *rpc.RpcError
}

type ContentUploaderWrapper struct {
	defaultTimeout time.Duration
	apiUrl         string
	serviceName    string
	baseWrapper    *wrappers.BaseWrapper
}

func NewContentUploaderWrapper(config boilerplate.WrapperConfig) IContentUploaderWrapper {
	timeout := 5 * time.Second

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

func (w ContentUploaderWrapper) UploadContentInternal(path string, contentType string, data string, apmTransaction *apm.Transaction, forceLog bool) chan *rpc.RpcError {
	resChan := make(chan *rpc.RpcError, 2)

	go func() {
		defer close(resChan)
		rpcInternalResponse := <-w.baseWrapper.SendRequestWithRpcResponseFromAnyService(fmt.Sprintf("%v/%v", w.apiUrl, path),
			"PUT",
			contentType,
			"upload content internal",
			data, map[string]string{}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)
		resChan <- rpcInternalResponse.Error
	}()

	return resChan
}
