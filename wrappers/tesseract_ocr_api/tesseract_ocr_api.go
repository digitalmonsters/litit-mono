package tesseract_ocr_api

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/wrappers"
	"github.com/rs/zerolog/log"
	"go.elastic.co/apm"
	"time"
)

type ITesseractOcrApiWrapper interface {
	RecognizeImageText(imageData []byte, languages []Language, hocrMode bool, trim bool, ctx context.Context,
		forceLog bool) chan wrappers.GenericResponseChan[RecognizeImageTextResponse]
}

//goland:noinspection GoNameStartsWithPackageName
type TesseractOcrApiWrapper struct {
	baseWrapper    *wrappers.BaseWrapper
	defaultTimeout time.Duration
	serviceApiUrl  string
	publicApiUrl   string
	serviceName    string
}

func NewTesseractOcrApiWrapper(config boilerplate.WrapperConfig) ITesseractOcrApiWrapper {
	timeout := 5 * time.Second

	if config.TimeoutSec > 0 {
		timeout = time.Duration(config.TimeoutSec) * time.Second
	}

	if len(config.ApiUrl) == 0 {
		config.ApiUrl = "http://tesseract-ocr-api"

		log.Warn().Msgf("Api Url is missing for TesseractOcrApi. Setting as default : %v", config.ApiUrl)
	}

	return &TesseractOcrApiWrapper{
		baseWrapper:    wrappers.GetBaseWrapper(),
		defaultTimeout: timeout,
		serviceApiUrl:  fmt.Sprintf("%v/rpc-service", common.StripSlashFromUrl(config.ApiUrl)),
		publicApiUrl:   common.StripSlashFromUrl(config.ApiUrl),
		serviceName:    "tesseract-ocr-api",
	}
}

func (w TesseractOcrApiWrapper) RecognizeImageText(imageData []byte, languages []Language, hocrMode bool, trim bool, ctx context.Context,
	forceLog bool) chan wrappers.GenericResponseChan[RecognizeImageTextResponse] {
	return wrappers.ExecuteRpcRequestAsync[RecognizeImageTextResponse](w.baseWrapper, w.serviceApiUrl,
		"RecognizeImageText", RecognizeImageTextRequest{
			ImageData: imageData,
			Languages: languages,
			HocrMode:  hocrMode,
			Trim:      trim,
		}, map[string]string{}, w.defaultTimeout, apm.TransactionFromContext(ctx), w.serviceName, forceLog)
}
