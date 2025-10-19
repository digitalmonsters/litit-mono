package tesseract_ocr_api

import (
	"context"
	"github.com/digitalmonsters/go-common/wrappers"
)

//goland:noinspection ALL
type TesseractOcrApiWrapperMock struct {
	RecognizeImageTextFn func(imageData []byte, languages []Language, hocrMode bool, trim bool, ctx context.Context,
		forceLog bool) chan wrappers.GenericResponseChan[RecognizeImageTextResponse]
}

func (m *TesseractOcrApiWrapperMock) RecognizeImageText(imageData []byte, languages []Language, hocrMode bool, trim bool,
	ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[RecognizeImageTextResponse] {
	return m.RecognizeImageTextFn(imageData, languages, hocrMode, trim, ctx, forceLog)
}

func GetMock() ITesseractOcrApiWrapper { // for compiler errors
	return &TesseractOcrApiWrapperMock{}
}
