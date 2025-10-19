package content_uploader

import (
	"go.elastic.co/apm"
)

type ContentUploaderWrapperMock struct {
	UploadContentInternalFn func(url string, contentType string, data []byte, apmTransaction *apm.Transaction, forceLog bool) chan error
}

func (m ContentUploaderWrapperMock) UploadContentInternal(url string, contentType string, data []byte, apmTransaction *apm.Transaction, forceLog bool) chan error {
	return m.UploadContentInternalFn(url, contentType, data, apmTransaction, forceLog)
}

func GetMock() IContentUploaderWrapper { // for compiler errors
	return &ContentUploaderWrapperMock{}
}
